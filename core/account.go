package core

import (
	"ablockchain/crypto"
	"ablockchain/trie"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

type Account struct {
	Address   string
	PublicKey []byte
	SecretKey []byte
	Balance   uint64
	Nonce     uint64
}

// StateDB 管理账户状态
type StateDB struct {
	trie *trie.Trie
	lock sync.RWMutex
}

func NewStateDB(dbPath string) (*StateDB, error) {
	trieDB, err := trie.NewTrie(dbPath)
	if err != nil {
		return nil, err
	}
	return &StateDB{trie: trieDB}, nil
}

// 创建新账户
func (s *StateDB) NewAccount() (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}

	privKeyBytes := crypto.FromECDSA(privKey)
	pubKeyBytes := crypto.FromECDSAPub(&privKey.PublicKey)
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()

	account := &Account{
		Address:   address,
		PublicKey: pubKeyBytes,
		SecretKey: privKeyBytes,
		Balance:   0,
		Nonce:     0,
	}

	// 存储账户到StateDB
	err = s.UpdateAccount(account)
	if err != nil {
		return nil, err
	}

	log.Printf("创建新账户，地址: %s\n", address)
	return account, nil
}

// 更新账户状态
func (s *StateDB) UpdateAccount(account *Account) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// RLP 编码账户数据
	data, err := rlp.EncodeToBytes(account)
	if err != nil {
		return err
	}
	return s.trie.Insert(account.Address, data)
}

// 查询账户，并解码 RLP
func (s *StateDB) GetAccount(address string) (*Account, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	data, err := s.trie.Get(address)
	if err != nil {
		return nil, false
	}

	var account Account
	if err := rlp.DecodeBytes(data, &account); err != nil {
		return nil, false
	}
	fmt.Printf("账户地址: %s, 账户余额: %d", account.Address, account.Balance)
	return &account, true
}

func (s *StateDB) Close() {
	s.trie.DB.Close()
}

// 打印 StateDB 中的所有账户及其余额
func (s *StateDB) PrintAccounts() {
	s.lock.RLock()
	defer s.lock.RUnlock()

	fmt.Println("\n账户列表:")
	iter := s.trie.NewIterator()

	for iter.Next() {
		var account Account
		if err := rlp.DecodeBytes(iter.Value(), &account); err != nil {
			log.Printf("账户解码失败: %v", err)
			continue
		}
		fmt.Printf("地址: %s | 余额: %d\n", account.Address, account.Balance)
	}
}
