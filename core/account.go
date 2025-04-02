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
	Address    string
	PublicKey  []byte //由于*ecdsa.PublicKey不支持rlp编码,必须使用[]byte
	PrivateKey []byte
	Balance    uint64
	Nonce      uint64
}

// StateDB 管理账户状态
type StateDB struct {
	trie *trie.Trie
	lock sync.RWMutex
}

// 打包签名和交易
type SignedTx struct {
	Tx   *Transaction
	Sign []byte
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
	// 存储账户到StateDB
	err = s.UpdateAccount(account) //RLP编码, 不支持*ecdsa.PrivateKey类型
	if err != nil {
		return nil, fmt.Errorf("存储账户到StateDB失败: %v", err)
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
	fmt.Printf("$$ 账户信息更新: \n")
	fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", account.Address, account.Balance, account.Nonce)
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
		fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", account.Address, account.Balance, account.Nonce)
	}
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
