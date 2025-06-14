package core

import (
	"ablockchain/cli"
	"ablockchain/config"
	"ablockchain/crypto"
	"ablockchain/event"
	"ablockchain/storage"
	"bytes"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
)

const BlockInterval time.Duration = 5 * time.Second
const MinTransactionsPerBlock int = 1 // 目前将区块交易数量固定为 MinTransactionsPerBlock
const MaxTransactionsPerBlock int = 10

type Blockchain struct {
	ChainID string
	DB      *storage.LevelDB
	TxPool  *TxPool
	StateDB *state.StateDB
	IndexDB *leveldb.DB // 用于保存索引:哈希->区块高度, 哈希->交易索引等
	//StateRoot []byte         // Merkle Patricia Tree的根哈希

	CurrentBlockHash []byte
	CurBlockNum      *big.Int
	NewBlockChan     chan *Block

	Noncelist map[string]uint64 // 管理本地交易的nonce，发送多笔交易时保证交易nonce的连续性
	Acclist   []string          // 记录账户列表（geth的statedb迭代器被隐藏，暂时无法遍历账户）
	Contrlist []string          // 记录合约地址
}

// TODO: 区块链侧可能也需要加一个计时器，当某个区块共识一直没好时，重新把这个区块发给共识模块
func NewBlockchain(cfg *cli.Config) (*Blockchain, error) {
	// path := "./block_storage"
	DBPath := cfg.DBPath
	db, err := storage.NewLevelDB(DBPath)
	if err != nil {
		return nil, err
	}

	txPool := NewTxPool()

	// 加载创世配置（需要在启动的目录下有一个创世文件
	// TODO 将genesis的路径作为一个可选参数放进cli里
	// default在项目目录下
	genensisConfig, err := config.LoadGenesisConfig("./genesis.json")
	if err != nil {
		return nil, err
	}
	log.Printf("加载创世配置……\n")

	stateDB := NewStateDB(DBPath + "_state")

	// 创建创世区块
	genesisBlock := NewGenesisBlock(genensisConfig.Difficulty)
	curBlockNum := big.NewInt(0)
	log.Printf("创建创世区块……\n")
	db.Put("0", genesisBlock)

	// // 计算初始 stateRoot
	// stateRoot := stateDB.trie.RootHash()
	indexdb, err := leveldb.OpenFile(cfg.DBPath+"/indexdb", nil)
	if err != nil {
		return nil, err
	}

	return &Blockchain{
		ChainID: cfg.ChainID,
		DB:      db,
		TxPool:  txPool,
		StateDB: stateDB,
		IndexDB: indexdb,
		//StateRoot: nil,
		//currentBlockHash: currentBlockHash,
		CurBlockNum:  curBlockNum,
		NewBlockChan: make(chan *Block, 10),
		Noncelist:    make(map[string]uint64),
	}, nil
}

// 开始一个异步的矿工进程
func (bc *Blockchain) StartMiner(miner string) {
	// 订阅交易消息
	txCh := event.Bus.Subscribe("TransactionMessage")

	// 监听交易事件
	go func() {
		for {
			select {
			case data := <-txCh:
				fmt.Print("\n##Miner开始验证交易##")
				signtx, ok := data.(*SignedTx)
				if !ok {
					log.Fatal("转换失败: 事件数据不是 *core.SignedTx 类型")
				}
				bc.handleTransaction(signtx) // 处理交易

			default:
				// 当交易池足够大时，打包区块
				if bc.TxPool.PendingSize() >= MinTransactionsPerBlock {
					bc.mineNewBlock(miner)
				}
				time.Sleep(BlockInterval)
			}
		}
	}()
}

// 处理收到的交易
func (bc *Blockchain) handleTransaction(signtx *SignedTx) {
	// 验证交易合法性
	valid, err := signtx.Tx.VerifySignature(signtx.Sign)
	if err != nil {
		log.Fatal("\n验证签名失败:", err)
	}
	if valid {
		fmt.Println("\n签名合法")
		signtx.Tx.PrintTransaction()
		// 交易加入交易池
		bc.addTx(signtx.Tx)
		fmt.Printf("交易池累计交易数：\n Pending: %d, Queued: %d\n", bc.TxPool.PendingSize(), bc.TxPool.QueuedSize())
	} else {
		log.Fatal("\n非法签名")
	}
}

// 将交易加入交易池，根据nonce判断加入Pending或Queued
func (bc *Blockchain) addTx(tx *Transaction) {
	maxNonce := bc.TxPool.GetMaxNonce(tx.From)
	// 若交易池中没有，要检查MPT中的nonce值（防止交易提交后，后面的交易nonce接不上）
	if maxNonce == 0 {
		from, ok := bc.GetAccState(tx.From)
		if ok {
			maxNonce = from.Nonce
		}
	}
	fmt.Printf("tx.Nonce：%d ； maxNonce：%d\n", tx.Nonce, maxNonce)
	if tx.Nonce == maxNonce+1 {
		bc.TxPool.AddPending(tx)
	} else {
		bc.TxPool.AddQueued(tx)
		return
	}
	// 尝试从 Queued 中连续推进 Pending
	nextNonce := tx.Nonce + 1
	for {
		nextTx := bc.TxPool.FindQueuedTxByNonce(tx.From, nextNonce)
		if nextTx == nil {
			break
		}
		bc.TxPool.AddPending(nextTx)
		bc.TxPool.RemoveQueued(nextTx)
		nextNonce++
	}
}

func (bc *Blockchain) mineNewBlock(miner string) (*Block, error) {
	txs := bc.TxPool.GetPendingTxs(MinTransactionsPerBlock)
	if len(txs) == 0 {
		return nil, fmt.Errorf("no transaction!")
	}
	// 避免打包已经提交过的区块
	merkleroot := CalculateMerkleRoot(txs)
	if bytes.Equal(merkleroot, bc.GetLatestMR()) {
		bc.TxPool.ClearPackedTxs(txs)
		fmt.Printf("打包的区块已上链\n")
		return nil, fmt.Errorf("重复区块")
	}

	// 创建新区块（该部分的difficulty需要进一步修改）
	header := NewBlockHeader(bc.CurrentBlockHash, uint64(2), new(big.Int).Add(bc.CurBlockNum, big.NewInt(1)), merkleroot, miner)
	block := NewBlock(header, txs)

	// bc.AddBlock(block)
	bc.NewBlockChan <- block // 将新区块发送到通道

	bc.TxPool.ClearPackedTxs(block.Transactions)
	fmt.Printf("打包区块，交易池：\n Pending: %d, Queued: %d\n", bc.TxPool.PendingSize(), bc.TxPool.QueuedSize())
	return block, nil
}

func (bc *Blockchain) AddBlock(block *Block) {
	str := fmt.Sprintf("%s", block.Header.Number)
	fmt.Println(str)
	block.PrintBlock()
	bc.ConfirmBlock(block)                       //修改账户状态MPT
	bc.DB.Put(str, block)                        //存入数据库区块高度->区块
	bc.StoreBlockIndex(block)                    // 添加索引:区块哈希->区块高度,方便查找
	bc.StoreTxIndex(block)                       // 添加交易索引
	bc.TxPool.ClearPackedTxs(block.Transactions) //清理交易池,防止重复打包
	fmt.Printf("上链，交易池：\n Pending: %d, Queued: %d\n", bc.TxPool.PendingSize(), bc.TxPool.QueuedSize())
	bc.CurBlockNum = block.Header.Number
	bc.CurrentBlockHash = block.Hash
}

// 返回上一个区块的默克尔根
func (bc *Blockchain) GetLatestMR() []byte {
	_, data, err := bc.DB.GetLatest()
	if err == nil {
		var block Block
		err = rlp.DecodeBytes(data, &block)
		if err == nil {
			return block.Header.MerkleRoot
		} else {
			fmt.Println("RLP decode error:", err)
			return nil
		}
	}
	fmt.Println("No data found.")
	return nil
}

func (bc *Blockchain) PrintLatest() {
	key, data, err := bc.DB.GetLatest()
	if err == nil {
		var value Block
		err = rlp.DecodeBytes(data, &value)
		if err == nil {
			fmt.Printf("Latest Block: %s\n", key)
			value.PrintBlock()
		} else {
			fmt.Println("RLP decode error:", err)
		}
	} else {
		fmt.Println("No data found.")
	}
}

func (bc *Blockchain) PrintAll() {
	allData, err := bc.DB.GetAll()
	if err == nil {
		fmt.Println("All Blocks:")
		for _, kv := range allData {
			var value Block
			err := rlp.DecodeBytes(kv.Value, &value)
			if err == nil {
				fmt.Printf("Block: %s\n", kv.Key)
				value.PrintBlock()
			} else {
				fmt.Println("RLP decode error:", err)
			}
		}
	} else {
		fmt.Println("No data found.")
	}
}

func NewTestBlockchain(DBPath string) (*Blockchain, error) {
	// path := "./block_storage"
	db, err := storage.NewLevelDB(DBPath)
	if err != nil {
		return nil, err
	}

	txPool := NewTxPool()

	stateDB := NewStateDB(DBPath + "_state")

	// 创建创世区块
	genesisBlock := NewGenesisBlock(2)
	curBlockNum := big.NewInt(0)
	log.Printf("创建创世区块……\n")
	db.Put("0", genesisBlock)

	// // 计算初始 stateRoot
	// stateRoot := stateDB.trie.RootHash()

	return &Blockchain{
		DB:      db,
		TxPool:  txPool,
		StateDB: stateDB,
		//StateRoot: nil,
		//currentBlockHash: currentBlockHash,
		CurBlockNum:  curBlockNum,
		NewBlockChan: make(chan *Block, 10),
		Noncelist:    make(map[string]uint64),
	}, nil
}

// 计算交易列表的 Merkle Root
func CalculateMerkleRoot(transactions []*Transaction) []byte {
	var txHashes [][]byte
	for _, tx := range transactions {
		txHashes = append(txHashes, tx.TxHash)
	}
	return crypto.ComputeMerkleRoot(txHashes)
}

// 读取编号从From到To编号的区块，返回列表
func (bc *Blockchain) GetBlocksInRange(from, to *big.Int) ([]*Block, error) {
	var blocks []*Block
	one := big.NewInt(1)
	i := new(big.Int).Set(from)

	for i.Cmp(to) <= 0 {
		key := i.String()
		var value Block
		err := bc.DB.Get(key, &value)
		if err != nil {
			return nil, fmt.Errorf("读取区块 %s 失败: %w", key, err)
		}
		blocks = append(blocks, &value)
		i.Add(i, one)
	}
	return blocks, nil
}
