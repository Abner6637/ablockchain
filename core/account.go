package core

import (
	"ablockchain/crypto"
	"ablockchain/trie"
	"crypto/ecdsa"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

type Account struct {
	Address    string
	PublicKey  *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
	Balance    uint64
	Nonce      uint64
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
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	account := &Account{
		Address:    address,
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
		Balance:    0,
		Nonce:      0,
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

// 对生成的交易进行签名
func (a *Account) SignTx(tx *Transaction) ([]byte, error) {
	encodetx, err := tx.EncodeTx() //rlp编码
	if err != nil {
		return nil, err
	}
	hashTx := crypto.GlobalHashAlgorithm.Hash(encodetx) //编码后计算hash
	return crypto.Sign(hashTx, a.PrivateKey)
}

func (tx *Transaction) VerifySignature(signature []byte) (bool, error) {
	encodedTx, err := tx.EncodeTx()
	if err != nil {
		return false, err
	}
	hashTx := crypto.GlobalHashAlgorithm.Hash(encodedTx)
	// 从哈希和签名恢复出公钥
	pubKey, err := crypto.SigToPub(hashTx, signature)
	if err != nil {
		return false, err
	}
	// 计算公钥对应的地址
	recoveredAddress := crypto.PubkeyToAddress(*pubKey).Hex()
	// 比对地址是否一致
	return recoveredAddress == tx.From, nil
}

// func (a *Account) SignMessage(msg *pbfttypes.Message) ([]byte, error) {
// 	// Sign message
// 	data, err := msg.PayloadNoSig()
// 	if err != nil {
// 		return nil, err
// 	}
// 	msg.Signature, err = c.Sign(data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return msg.Signature, err
// }

// // 通过未经哈希的原数据和签名得到签名所用的公钥，再通过公钥得到签名地址
// func GetSignatureAddress(data []byte, sig []byte) (common.Address, error) {
// 	hashData := crypto.GlobalHashAlgorithm.Hash(data)

// 	pubkey, err := crypto.SigToPub(hashData, sig)
// 	if err != nil {
// 		return common.Address{}, err
// 	}
// 	return crypto.PubkeyToAddress(*pubkey), nil
// }

// func VerifySignature(msg *pbfttypes.Message) error {
// 	payloadNoSig, err := msg.PayloadNoSig()
// 	if err != nil {
// 		return err
// 	}

// 	signerAddress, err := GetSignatureAddress(payloadNoSig, msg.Signature)

// 	// 比较签名地址和消息中的地址参数（即发送消息的地址）是否一致
// 	if !bytes.Equal(signerAddress.Bytes(), msg.Address) {
// 		return errors.New("invaid signer")
// 	}
// 	return nil
// }
