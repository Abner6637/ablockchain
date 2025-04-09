package core

import (
	"ablockchain/crypto"
	"ablockchain/trie/mpt"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

var oribalance = uint64(100)
var privkeyaddr = "./privkey"

type Account struct {
	PublicKey  []byte //由于*ecdsa.PublicKey不支持rlp编码,必须使用[]byte
	PrivateKey []byte

	State *AccState
}

// 记录账户状态，用于保存到MPT
type AccState struct {
	Address string
	Balance uint64
	Nonce   uint64
}

type AccountKeyFile struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"` // hex or base64 encoded
	PublicKey  string `json:"public_key"`
}

func SaveAccountToFile(account *Account, dir string) error {
	keyFile := AccountKeyFile{
		Address:    account.State.Address,
		PrivateKey: hex.EncodeToString(account.PrivateKey),
		PublicKey:  hex.EncodeToString(account.PublicKey),
	}

	data, err := json.MarshalIndent(keyFile, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化账户失败: %v", err)
	}

	// 文件名使用地址
	fileName := filepath.Join(dir, fmt.Sprintf("%s.json", account.State.Address))
	err = os.WriteFile(fileName, data, 0600)
	if err != nil {
		return fmt.Errorf("写入账户文件失败: %v", err)
	}
	return nil
}

func loadAccountKeysFromFile(address, dir string) ([]byte, []byte, error) {
	file := filepath.Join(dir, fmt.Sprintf("%s.json", address))
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("读取账户文件失败: %v", err)
	}

	var keyFile AccountKeyFile
	err = json.Unmarshal(data, &keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("解析账户文件失败: %v", err)
	}

	privKey, err := hex.DecodeString(keyFile.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("解码私钥失败: %v", err)
	}
	pubKey, err := hex.DecodeString(keyFile.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("解码公钥失败: %v", err)
	}

	return privKey, pubKey, nil
}

func (s *StateDB) GetAccount(address string) (*Account, error) {
	accState, ok := s.GetAccState(address)
	if !ok {
		return nil, fmt.Errorf("找不到账户状态: %s", address)
	}

	privKey, pubKey, err := loadAccountKeysFromFile(address, privkeyaddr)
	if err != nil {
		return nil, fmt.Errorf("加载账户密钥失败: %v", err)
	}

	return &Account{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		State:      accState,
	}, nil
}

// StateDB 管理账户状态(内存操作，提交到Trie后才是持久化存储)
type StateDB struct {
	trie *mpt.Trie
	lock sync.RWMutex

	stateObjects      map[string]*AccState // 缓存账户对象（address -> Account）
	stateObjectsDirty map[string]struct{}  // 跟踪被修改的账户地址
}

// 打包 签名和交易
type SignedTx struct {
	Tx   *Transaction
	Sign []byte
}

func NewStateDB(dbPath string) (*StateDB, error) {
	trieDB, err := mpt.NewTrie(dbPath)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		trie:              trieDB,
		stateObjects:      make(map[string]*AccState),
		stateObjectsDirty: make(map[string]struct{}),
	}, nil
}

// 创建新账户(暂存在内存中，只有进行交易后才加入MPT)
func (s *StateDB) NewAccount() (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	privKeyBytes := crypto.FromECDSA(privKey)
	pubKeyBytes := crypto.FromECDSAPub(&privKey.PublicKey)
	accstate := &AccState{
		Address: address,
		Balance: oribalance,
		Nonce:   0,
	}
	account := &Account{
		PrivateKey: privKeyBytes,
		PublicKey:  pubKeyBytes,
		State:      accstate,
	}
	// 账户标记
	// err = s.UpdateAccount(account) //RLP编码, 不支持*ecdsa.PrivateKey类型
	// if err != nil {
	// 	return nil, fmt.Errorf("存储账户到StateDB失败: %v", err)
	// }
	// 存到mpt和leveldb
	// s.Commit()

	s.stateObjects[account.State.Address] = accstate //暂存到内存中，不提交（产生交易后才提交）

	log.Printf("创建新账户，地址: %s\n", address)
	// 保存到本地

	_ = os.MkdirAll(privkeyaddr, 0700)
	err = SaveAccountToFile(account, privkeyaddr)
	if err != nil {
		return nil, fmt.Errorf("保存账户文件失败: %v", err)
	}

	return account, nil
}

// 更新账户状态(StateDB),避免重复提交
func (s *StateDB) UpdateAccState(accstate *AccState) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.stateObjects[accstate.Address] = accstate
	s.stateObjectsDirty[accstate.Address] = struct{}{}

	fmt.Printf("$$ 账户信息更新: 地址: %s | 余额: %d | Nonce: %d\n", accstate.Address, accstate.Balance, accstate.Nonce)
	return nil
}

// 查询账户，并解码 RLP
func (s *StateDB) GetAccState(address string) (*AccState, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// 优先在statedb中查找
	if accs, ok := s.stateObjects[address]; ok {
		return accs, true
	}

	// 在mpt中查找
	data, err := s.trie.Get(address)
	if err != nil || data == nil {
		return nil, false
	}
	var accstate AccState
	if err := rlp.DecodeBytes(data, &accstate); err != nil {
		return nil, false
	}

	s.stateObjects[address] = &accstate
	return &accstate, true
}

// 打印 StateDB 中的所有账户及其余额
func (s *StateDB) PrintAccounts() {
	fmt.Println("=== StateDB中的账户 ===")
	for _, accs := range s.stateObjects {
		fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", accs.Address, accs.Balance, accs.Nonce)
	}

	fmt.Println("=== MPT中的账户 ===")
	s.TraverseAccState(func(key string, accs *AccState) {
		fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", accs.Address, accs.Balance, accs.Nonce)
	})
}

func (s *StateDB) Close() {
	s.trie.DB.Close()
}

// 对生成的交易进行签名, 返回rlp编码的SignedTx结构体
func (a *Account) SignTx(tx *Transaction) (*SignedTx, error) {
	encodetx, err := tx.EncodeTx()
	if err != nil {
		return nil, err
	}
	hashTx := crypto.GlobalHashAlgorithm.Hash(encodetx) //编码后计算hash
	privk, err := crypto.ToECDSA(a.PrivateKey)
	if err != nil {
		return nil, err
	}
	sign, err := crypto.Sign(hashTx, privk)
	if err != nil {
		return nil, err
	}
	signedTx := &SignedTx{
		Tx:   tx,
		Sign: sign,
	}
	return signedTx, nil
}

func DecodeSignTx(data []byte) (*SignedTx, error) {
	var signtx SignedTx
	err := rlp.DecodeBytes(data, &signtx)
	if err != nil {
		log.Fatal("Failed to decode RLP SignedTx:", err)
		return nil, err
	}
	return &signtx, nil
}

// 区块上链后，改变账户状态
// TODO: 验证交易与账户的nonce信息，是否按顺序进行
func (s *StateDB) ConfirmBlock(block *Block) error {
	for _, tx := range block.Transactions {
		from, ok := s.GetAccState(tx.From)
		if ok {
			from.Balance -= tx.Value
			s.UpdateAccState(from)
		}
		to, ok := s.GetAccState(tx.To)
		if ok {
			to.Balance += tx.Value
			s.UpdateAccState(to)
		} else { //若mpt中没有，说明是新建的账户
			accs := &AccState{
				Address: tx.To,
				Balance: tx.Value,
				Nonce:   0,
			}
			s.UpdateAccState(accs)
		}
		//TODO:判断用户是否存在，前提：statedb的同步
	}
	s.Commit() //更新MPT
	return nil
}

// 将account插入mpt，并将acc和mpt都提交到leveldb
func (s *StateDB) Commit() ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for addr := range s.stateObjectsDirty {
		account := s.stateObjects[addr]
		data, err := rlp.EncodeToBytes(account)
		if err != nil {
			return nil, err
		}
		s.trie.Insert(addr, data)
	}

	// 提交到 LevelDB
	root, err := s.trie.Commit()
	if err != nil {
		return nil, err
	}

	// 清除状态
	s.stateObjectsDirty = make(map[string]struct{})
	return root, nil
}

// 用于遍历
func (s *StateDB) TraverseAccState(fn func(key string, accs *AccState)) {
	if s.trie.DB == nil {
		return
	}
	s.trie.DB.ForEach(func(key string, value []byte) {
		accs, ok := s.GetAccState(key)
		if ok && accs.Address != "" {
			fn(accs.Address, accs)
		}
	})
}
