package core

import (
	"ablockchain/cli"
	"ablockchain/config"
	"ablockchain/storage"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

const BlockInterval time.Duration = 5 * time.Second
const MinTransactionsPerBlock int = 2
const MaxTransactionsPerBlock int = 10

type Blockchain struct {
	DB        *storage.LevelDB
	TxPool    *TxPool
	StateDB   *StateDB // 使用Merkle Patricia Tree来存储账户状态
	StateRoot []byte   // Merkle Patricia Tree的根哈希

	CurrentBlockHash []byte
	CurBlockNum      uint64
	NewBlockChan     chan *Block
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
	genensisConfig, err := config.LoadGenesisConfig("./genesis.json")
	if err != nil {
		return nil, err
	}
	log.Printf("加载创世配置……\n")

	stateDB, err := NewStateDB(DBPath + "_state")
	if err != nil {
		return nil, err
	}

	// 创建创世区块
	genesisBlock := NewGenesisBlock(genensisConfig.Difficulty)
	curBlockNum := uint64(0)
	log.Printf("创建创世区块……\n")
	db.Put("0", genesisBlock)

	// 计算初始 stateRoot
	stateRoot := stateDB.trie.RootHash()

	return &Blockchain{
		DB:        db,
		TxPool:    txPool,
		StateDB:   stateDB,
		StateRoot: stateRoot,
		//currentBlockHash: currentBlockHash,
		CurBlockNum:  curBlockNum,
		NewBlockChan: make(chan *Block, 10),
	}, nil
}

// 开始一个异步的miner进程
func (bc *Blockchain) StartMiner() {
	go func() {
		for {
			if bc.TxPool.PendingSize() >= MinTransactionsPerBlock {
				bc.mineNewBLock()
			}
			time.Sleep(BlockInterval)
		}
	}()
}

func (bc *Blockchain) mineNewBLock() (*Block, error) {
	txs := bc.TxPool.GetTxs()
	if len(txs) == 0 {
		return nil, fmt.Errorf("no transaction!")
	}

	// 创建新区块（该部分的difficulty需要进一步修改）
	header := NewBlockHeader(bc.CurrentBlockHash, uint64(1), bc.CurBlockNum+1)
	block := NewBlock(header, txs)

	// bc.AddBlock(block)
	bc.NewBlockChan <- block // 将新区块发送到通道

	bc.TxPool.ClearPackedTxs(block.Transactions)
	return block, nil
}

func (bc *Blockchain) AddBlock(block *Block) {
	str := fmt.Sprintf("%d", block.Header.Number)
	fmt.Println(str)
	block.PrintBlock()
	bc.DB.Put(str, block)
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

	// 加载创世配置（需要在启动的目录下有一个创世文件
	genensisConfig, err := config.LoadGenesisConfig("./genesis.json")
	if err != nil {
		return nil, err
	}
	log.Printf("加载创世配置……\n")

	stateDB, err := NewStateDB(DBPath + "_state")
	if err != nil {
		return nil, err
	}

	// 创建创世区块
	genesisBlock := NewGenesisBlock(genensisConfig.Difficulty)
	curBlockNum := uint64(0)
	log.Printf("创建创世区块……\n")
	db.Put("0", genesisBlock)

	// 计算初始 stateRoot
	stateRoot := stateDB.trie.RootHash()

	return &Blockchain{
		DB:        db,
		TxPool:    txPool,
		StateDB:   stateDB,
		StateRoot: stateRoot,
		//currentBlockHash: currentBlockHash,
		CurBlockNum:  curBlockNum,
		NewBlockChan: make(chan *Block, 10),
	}, nil
}
