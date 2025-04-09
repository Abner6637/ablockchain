package core

import (
	"ablockchain/crypto"
	"ablockchain/trie/mpt"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"
)

var oribalance = uint64(100)
var privkeyaddr = "./privkey"

// StateDB 管理账户状态(内存操作，提交到Trie后才是持久化存储)
type StateDB struct {
	trie *mpt.Trie
	lock sync.RWMutex

	stateObjects      map[string]*AccState // 缓存账户对象（address -> Account）
	stateObjectsDirty map[string]struct{}  // 跟踪被修改的账户地址
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

// 返回完整账户结构（包括密钥）
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
