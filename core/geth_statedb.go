package core

import (
	"ablockchain/crypto"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

var oribalance = uint64(1000)
var privkeyaddr = "./privkey"
var reward = uint64(10)

// StateDB 管理账户状态(内存操作，提交到Trie后才是持久化存储)
// type StateDB struct {
// 	trie *mpt.Trie
// 	lock sync.RWMutex

// 	stateObjects      map[string]*AccState // 缓存账户对象（address -> Account）
// 	stateObjectsDirty map[string]struct{}  // 跟踪被修改的账户地址
// 	Noncelist         map[string]uint64    // 管理本地交易的nonce，发送多笔交易时保证交易nonce的连续性
// }

func NewStateDB(dbPath string) *state.StateDB {
	dir := dbPath
	ldb, err := leveldb.New(dir, 256, 512, "", false)
	if err != nil {
		fmt.Printf("Failed to create leveldb: %v", err)
	}

	db := rawdb.NewDatabase(ldb)
	triedb := triedb.NewDatabase(db, nil)

	statedb, err := state.New(types.EmptyRootHash, state.NewDatabase(triedb, nil))
	if err != nil {
		panic(err)
	}

	return statedb
}

func (bc *Blockchain) AddAccList(account string) {
	// 检查账户是否已存在
	for _, acc := range bc.Acclist {
		if acc == account {
			return // 账户已存在，直接返回
		}
	}

	// 账户不存在，添加到列表
	bc.Acclist = append(bc.Acclist, account)
}

// // 返回完整账户结构（包括密钥）
// func (s *StateDB) GetAccount(address string) (*Account, error) {
// 	accState, ok := s.GetAccState(address)
// 	if !ok {
// 		return nil, fmt.Errorf("找不到账户状态: %s", address)
// 	}

// 	privKey, err := loadAccountKeysFromFile(address, privkeyaddr)
// 	if err != nil {
// 		return nil, fmt.Errorf("加载账户密钥失败: %v", err)
// 	}

// 	return &Account{
// 		PrivateKey: privKey,
// 		State:      accState,
// 	}, nil
// }

// 创建新账户(暂存在内存中，只有进行交易后才加入MPT)
func (bc *Blockchain) NewAccount() (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	privKeyBytes := crypto.FromECDSA(privKey)
	accstate := &AccState{
		Address: address,
		Balance: 0,
		Nonce:   0,
		Code:    nil,
	}
	account := &Account{
		PrivateKey: privKeyBytes,
		State:      accstate,
	}

	//暂存到内存中，不提交（产生交易后才提交）
	addr := common.HexToAddress(address)
	if bc.StateDB.Exist(addr) {
		return nil, fmt.Errorf("Account should not exist initially")
	}

	bc.StateDB.CreateAccount(addr)
	if !bc.StateDB.Exist(addr) {
		return nil, fmt.Errorf("Account should exist after creation")
	}
	if !bc.StateDB.Empty(addr) {
		return nil, fmt.Errorf("New account should be empty")
	}

	//s.stateObjects[account.State.Address] = accstate

	log.Printf("创建新账户，地址: %s\n", address)
	bc.AddAccList(address)
	// 保存到本地
	_ = os.MkdirAll(privkeyaddr, 0700)
	err = SaveAccountToFile(account, privkeyaddr)
	if err != nil {
		return nil, fmt.Errorf("保存账户文件失败: %v", err)
	}

	return account, nil
}

// 更新账户状态(StateDB),避免重复提交
func (bc *Blockchain) UpdateAccState(accstate *AccState) {

	// s.stateObjects[accstate.Address] = accstate
	// s.stateObjectsDirty[accstate.Address] = struct{}{}

	addr := common.HexToAddress(accstate.Address)
	if !bc.StateDB.Exist(addr) {
		bc.StateDB.CreateAccount(addr)
		bc.AddAccList(accstate.Address)
	}
	bc.StateDB.SetBalance(addr, uint256.NewInt(accstate.Balance), tracing.BalanceChangeUnspecified)
	bc.StateDB.SetNonce(addr, accstate.Nonce, tracing.NonceChangeUnspecified)
	if accstate.Code != nil {
		bc.StateDB.SetCode(addr, accstate.Code)

		fmt.Printf("$$ 合约账户: 地址: %s | 余额: %d | Nonce: %d\n", accstate.Address, accstate.Balance, accstate.Nonce)
		fmt.Printf("$$ 合约代码 %s: \n", hex.EncodeToString(accstate.Code))
	} else {
		fmt.Printf("$$ 账户信息更新: 地址: %s | 余额: %d | Nonce: %d\n", accstate.Address, accstate.Balance, accstate.Nonce)
	}

}

// 查询账户
func (bc *Blockchain) GetAccState(address string) (*AccState, bool) {
	addr := common.HexToAddress(address)
	if bc.StateDB.Exist(addr) {
		var accstate AccState
		accstate.Address = address
		accstate.Balance = bc.StateDB.GetBalance(addr).Uint64()
		accstate.Nonce = bc.StateDB.GetNonce(addr)

		return &accstate, true

	}
	return nil, false
}

func (bc *Blockchain) PrintAllAccounts() {
	if len(bc.Acclist) == 0 {
		fmt.Println("账户列表为空")
		return
	}
	fmt.Println("====================================")

	for _, address := range bc.Acclist {
		accState, exists := bc.GetAccState(address)
		if !exists {
			fmt.Printf("账户 %s 不存在于状态数据库中\n", address)
			continue
		}
		fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", accState.Address, accState.Balance, accState.Nonce)
	}
}

// // 查询账户 StateDB
// func (s *StateDB) GetAccStateFromStateDB(address string) (*AccState, bool) {
// 	s.lock.RLock()
// 	defer s.lock.RUnlock()

// 	if accs, ok := s.stateObjects[address]; ok {
// 		return accs, true
// 	}

// 	return nil, false
// }

// // 查询账户 MPT
// func (s *StateDB) GetAccStateFromMPT(address string) (*AccState, bool) {
// 	s.lock.RLock()
// 	defer s.lock.RUnlock()

// 	data, err := s.trie.Get(address)
// 	if err != nil || data == nil {
// 		return nil, false
// 	}
// 	var accstate AccState
// 	if err := rlp.DecodeBytes(data, &accstate); err != nil {
// 		return nil, false
// 	}

// 	s.stateObjects[address] = &accstate
// 	return &accstate, true
// }

// 验证交易是否合法（付款账户是否存在？余额是否充足？nonce是否一致）
// func (s *StateDB) VeruifyTx(tx *Transaction) bool {
// 	acc, ok := s.GetAccState(tx.From)
// 	if !ok {
// 		fmt.Printf("付款账户不存在\n")
// 		return false
// 	}
// 	if acc.Balance < tx.Value {
// 		fmt.Printf("付款账户余额不足\n")
// 		return false
// 	}
// 	if acc.Nonce != tx.Nonce {
// 		fmt.Printf("Nonce不匹配\n")
// 		return false
// 	}

// 	return true
// }

// 打印 StateDB 中的所有账户及其余额

// func (bc *Blockchain) PrintAccounts() {
// 	fmt.Println("=== StateDB中的账户 ===")
// 	for _, accs := range bc.StateDB.stateObjects {
// 		fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", accs.Address, accs.Balance, accs.Nonce)
// 	}

// 	// fmt.Println("=== MPT中的账户 ===")
// 	// s.TraverseAccState(func(key string, accs *AccState) {
// 	// 	fmt.Printf("地址: %s | 余额: %d | Nonce: %d\n", accs.Address, accs.Balance, accs.Nonce)
// 	// })
// }

// func (s *StateDB) Close() {
// 	s.trie.DB.Close()
// }

// 区块上链后，根据区块信息改变账户状态，并提交至MPT
// TODO: 验证交易与账户的nonce信息，是否按顺序进行
func (bc *Blockchain) ConfirmBlock(block *Block) error {
	for _, tx := range block.Transactions {
		from, ok := bc.GetAccState(tx.From)
		if ok {
			// TODO: 是否需要在这里验证nonce的顺序？
			// if from.Nonce+1 == tx.Nonce {
			// 	from.Balance -= tx.Value
			// 	from.Nonce += 1
			// 	s.UpdateAccState(from)
			// } else {
			// 	return nil
			// }
			from.Balance -= tx.Value
			from.Nonce += 1
			bc.UpdateAccState(from)
		} else { // ###########实际情况下不应该记录from账户，因为初始余额为0不可能发起交易##### 正确做法是触发同步
			accs := &AccState{
				Address: tx.From,
				Balance: oribalance - tx.Value,
				Nonce:   1,
			}
			bc.UpdateAccState(accs)
		}
		to, ok := bc.GetAccState(tx.To)
		if ok {
			to.Balance += tx.Value
			bc.UpdateAccState(to)
		} else { //若mpt中没有，说明是新建的账户
			accs := &AccState{
				Address: tx.To,
				Balance: oribalance + tx.Value,
				Nonce:   0,
			}
			bc.UpdateAccState(accs)
		}
		miner, ok := bc.GetAccState(block.Header.Coinbase)
		if ok {
			miner.Balance += reward
			bc.UpdateAccState(miner)
		} else { //若mpt中没有，说明是新建的账户
			accs := &AccState{
				Address: block.Header.Coinbase,
				Balance: oribalance + reward,
				Nonce:   0,
			}
			bc.UpdateAccState(accs)
		}
		//TODO:判断用户是否存在，前提：statedb的同步
	}
	bc.Commit() //更新MPT
	return nil
}

// 将account插入mpt，并将acc和mpt都提交到leveldb
func (bc *Blockchain) Commit() ([]byte, error) {

	root, err := bc.StateDB.Commit(bc.CurBlockNum.Uint64(), false, false)
	if err != nil {
		return nil, err
	}
	return root.Bytes(), nil
}

// // 用于遍历
// func (s *StateDB) TraverseAccState(fn func(key string, accs *AccState)) {
// 	if s.trie.DB == nil {
// 		return
// 	}
// 	s.trie.DB.ForEach(func(key string, value []byte) {
// 		accs, ok := s.GetAccState(key)
// 		if ok && accs.Address != "" {
// 			fn(accs.Address, accs)
// 		}
// 	})
// }
