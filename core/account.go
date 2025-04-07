package core

import (
	"ablockchain/crypto"
	"ablockchain/trie/mpt"

	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

type Account struct {
	Address    string
	PublicKey  []byte //由于*ecdsa.PublicKey不支持rlp编码,必须使用[]byte
	PrivateKey []byte
	Balance    uint64
	Nonce      uint64
}

// StateDB 管理账户状态
type StateDB struct {
	trie *mpt.Trie
	lock sync.RWMutex

	stateObjects      map[string]*Account // 缓存账户对象（address -> Account）
	stateObjectsDirty map[string]struct{} // 跟踪被修改的账户地址
}

// 打包签名和交易
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
		stateObjects:      make(map[string]*Account),
		stateObjectsDirty: make(map[string]struct{}),
	}, nil
}

// 创建新账户
func (s *StateDB) NewAccount() (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	privKeyBytes := crypto.FromECDSA(privKey)
	pubKeyBytes := crypto.FromECDSAPub(&privKey.PublicKey)
	account := &Account{
		Address:    address,
		PrivateKey: privKeyBytes,
		PublicKey:  pubKeyBytes,
		Balance:    100,
		Nonce:      0,
	}
	// 存储账户到内存（脏标记）
	err = s.UpdateAccount(account) //RLP编码, 不支持*ecdsa.PrivateKey类型
	if err != nil {
		return nil, fmt.Errorf("存储账户到StateDB失败: %v", err)
	}
	// 存到mpt和leveldb
	s.Commit()

	log.Printf("创建新账户，地址: %s, 账户数：%d\n", address, s.trie.DB.Count)
	return account, nil
}

// 更新账户状态(内存)
func (s *StateDB) UpdateAccount(account *Account) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.stateObjects[account.Address] = account
	s.stateObjectsDirty[account.Address] = struct{}{}

	fmt.Printf("$$ 账户信息更新: 地址: %s | 余额: %d | Nonce: %d\n", account.Address, account.Balance, account.Nonce)
	return nil
}

// 查询账户，并解码 RLP
func (s *StateDB) GetAccount(address string) (*Account, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if acc, ok := s.stateObjects[address]; ok {
		return acc, true
	}

	data, err := s.trie.Get(address)
	if err != nil || data == nil {
		return nil, false
	}
	var account Account
	if err := rlp.DecodeBytes(data, &account); err != nil {
		return nil, false
	}

	s.stateObjects[address] = &account
	return &account, true
}

// 打印 StateDB 中的所有账户及其余额
func (s *StateDB) PrintAccounts() {
	fmt.Println("=== 内存中账户 ===")
	for _, acc := range s.stateObjects {
		fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", acc.Address, acc.Balance, acc.Nonce)
	}

	fmt.Println("=== LevelDB中提交的账户 ===")
	s.TraverseAccounts(func(key string, account *Account) {
		fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", account.Address, account.Balance, account.Nonce)
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
		from, ok := s.GetAccount(tx.From)
		if ok {
			from.Balance -= tx.Value
			from.Nonce += 1
			s.UpdateAccount(from)
		}
		to, ok := s.GetAccount(tx.To)
		if ok {
			to.Balance += tx.Value
			s.UpdateAccount(to)
		}
		//TODO:判断用户是否存在，前提：statedb的同步
	}
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
func (s *StateDB) TraverseAccounts(fn func(key string, account *Account)) {
	if s.trie.DB == nil {
		return
	}
	fmt.Println(">>> 遍历 LevelDB 内容中账户信息：")

	s.trie.DB.ForEach(func(key string, value []byte) {
		account, ok := s.GetAccount(key)
		if ok && account.Address != "" {
			fn(account.Address, account)
		}
	})
}
