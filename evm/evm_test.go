package evm

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

// 两个合约相互调用
func TestContractExecution(t *testing.T) {
	// 1. 准备测试合约
	a := program.New()
	b := program.New()

	aAddr := common.HexToAddress("0xff0a")
	bAddr := common.HexToAddress("0xff0b")

	// 合约A: 无限循环调用合约B
	//_, dest := a.Jumpdest()
	a.Call(nil, bAddr, nil, 0, 0, 0, 0)
	//a.Jump(dest)

	// 合约B: 回调调用者
	b.Op(vm.PC)              // 获取0值(out size)
	b.Op(vm.DUP1, vm.DUP1)   // out offset, insize
	b.Op(vm.DUP1, vm.DUP1)   // inoffset, value
	b.Op(vm.ADDRESS, vm.GAS) // address, gas
	b.Op(vm.CALL)            // 执行调用

	// 2. 设置初始状态
	alloc := make(types.GenesisAlloc)
	alloc[aAddr] = types.Account{
		Nonce:   0,
		Code:    a.Bytes(),
		Balance: big.NewInt(0xffffffff),
	}
	alloc[bAddr] = types.Account{
		Nonce:   0,
		Code:    b.Bytes(),
		Balance: big.NewInt(0xffffffff),
	}

	// 3. 验证初始状态JSON序列化
	outp, err := json.MarshalIndent(alloc, "", "  ")
	assert.NoError(t, err, "JSON marshaling failed")
	t.Logf("Initial state:\n%s", outp)

	// 4. 准备执行环境
	statedb := StateDBWithAlloc(alloc)
	sender := common.BytesToAddress([]byte("sender"))
	statedb.CreateAccount(sender)

	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    10000000,
		Difficulty:  big.NewInt(0x200000),
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			EIP150Block:         new(big.Int),
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
		},
	}

	// 5. 执行测试并测量时间
	t.Run("First execution with possible JIT compilation", func(t *testing.T) {
		start := time.Now()
		_, _, err := runtime.Call(aAddr, nil, &runtimeConfig)
		elapsed := time.Since(start)

		assert.NoError(t, err, "First execution failed")
		t.Logf("First execution time: %v", elapsed)
	})

	t.Run("Subsequent execution without compilation overhead", func(t *testing.T) {
		start := time.Now()
		_, _, err := runtime.Call(aAddr, nil, &runtimeConfig)
		elapsed := time.Since(start)

		assert.NoError(t, err, "Second execution failed")
		t.Logf("Second execution time: %v", elapsed)
	})
}

// get()，直接返回42（2a）
func TestSolidityBytecodeExecution(t *testing.T) {
	// Remix 编译后的 runtime 字节码（删除 0x 前缀）
	codeHex := "6080604052348015600f57600080fd5b506004361060285760003560e01c80636d4ce63c14602d575b600080fd5b60336047565b604051603e91906067565b60405180910390f35b6000602a905090565b6000819050919050565b6061816050565b82525050565b6000602082019050607a6000830184605a565b9291505056fea2646970667358221220857e54f745071af59d359e42dfacba868b7c6695cb"
	code, err := hex.DecodeString(codeHex)
	assert.NoError(t, err, "Failed to decode bytecode")

	contractAddr := common.HexToAddress("0xdeadbeef")
	sender := common.HexToAddress("0xaaaaaa")

	// 初始化账户状态
	alloc := make(types.GenesisAlloc)
	alloc[contractAddr] = types.Account{
		Code:    code,
		Nonce:   0,
		Balance: big.NewInt(1e18),
	}
	alloc[sender] = types.Account{
		Nonce:   0,
		Balance: big.NewInt(1e18),
	}

	// 初始化状态数据库
	statedb := StateDBWithAlloc(alloc)

	// EVM 运行时配置
	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    10_000_000,
		Difficulty:  big.NewInt(1),
		BlockNumber: big.NewInt(1),
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      big.NewInt(0),
			ByzantiumBlock:      big.NewInt(0),
			ConstantinopleBlock: big.NewInt(0),
			EIP150Block:         big.NewInt(0),
			EIP155Block:         big.NewInt(0),
			EIP158Block:         big.NewInt(0),
			PetersburgBlock:     big.NewInt(0),
			IstanbulBlock:       big.NewInt(0),
		},
	}

	// 选择器：keccak256("get()")[:4] = 0x6d4ce63c（来自字节码中的函数选择器）
	data := common.FromHex("6d4ce63c")

	t.Run("Solidity bytecode execution", func(t *testing.T) {
		start := time.Now()
		output, gasUsed, err := runtime.Call(contractAddr, data, &runtimeConfig)
		elapsed := time.Since(start)

		assert.NoError(t, err, "Execution failed")
		t.Logf("Execution succeeded in %v, gas used: %d", elapsed, gasUsed)
		t.Logf("Output: 0x%x", output)
	})
}

// 存储100，读取，比较
func TestStorageContract(t *testing.T) {
	// 从 solc 编译器获得的 runtime 字节码（仅部署后部分）
	bytecode := "608060405234801561001057600080fd5b50600436106100365760003560e01c80632e64cec11461003b5780636057361d14610059575b600080fd5b610043610075565b60405161005091906100a1565b60405180910390f35b610073600480360381019061006e91906100ed565b61007e565b005b60008054905090565b8060008190555050565b6000819050919050565b61009b81610088565b82525050565b60006020820190506100b66000830184610092565b92915050565b600080fd5b6100ca81610088565b81146100d557600080fd5b50565b6000813590506100e7816100c1565b92915050565b600060208284031215610103576101026100bc565b5b6000610111848285016100d8565b9150509291505056fea26469706673582212207bbff281ae8c9b6ea56e4da77beab4556e6bc57e474bf6d5e5b2e37f" // 这里填入 Storage 的实际 runtime 字节码

	code, err := hex.DecodeString(bytecode)
	assert.NoError(t, err, "Failed to decode bytecode")

	storageABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"name":"store","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"retrieve","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`))
	assert.NoError(t, err, "Failed to parse ABI")

	contractAddr := common.HexToAddress("0xdeadbeef")
	sender := common.HexToAddress("0xaaaaaa")

	alloc := make(types.GenesisAlloc)
	alloc[sender] = types.Account{
		Nonce:   0,
		Balance: big.NewInt(1e18),
	}

	statedb := StateDBWithAlloc(alloc)
	statedb.SetCode(contractAddr, code)

	runtimeCfg := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    10_000_000,
		BlockNumber: big.NewInt(1),
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      big.NewInt(0),
			ByzantiumBlock:      big.NewInt(0),
			ConstantinopleBlock: big.NewInt(0),
			EIP150Block:         big.NewInt(0),
			EIP155Block:         big.NewInt(0),
			EIP158Block:         big.NewInt(0),
			PetersburgBlock:     big.NewInt(0),
			IstanbulBlock:       big.NewInt(0),
		},
	}

	t.Run("Store and Retrieve", func(t *testing.T) {
		// 调用 store(100)
		input, err := storageABI.Pack("store", big.NewInt(100))
		assert.NoError(t, err, "ABI pack store failed")

		_, _, err = runtime.Call(contractAddr, input, &runtimeCfg)
		assert.NoError(t, err, "Call to store failed")

		// 调用 retrieve()
		input, err = storageABI.Pack("retrieve")
		assert.NoError(t, err, "ABI pack retrieve failed")

		output, _, err := runtime.Call(contractAddr, input, &runtimeCfg)
		assert.NoError(t, err, "Call to retrieve failed")

		// 解码返回值
		var result *big.Int
		err = storageABI.UnpackIntoInterface(&result, "retrieve", output)
		assert.NoError(t, err, "Unpack retrieve result failed")

		assert.Equal(t, big.NewInt(100), result, "retrieve() should return 100")
	})
}
