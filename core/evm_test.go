package core

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// 测试辅助函数 - 创建测试区块链
func setupTestBlockchain(t *testing.T) *Blockchain {
	bc, err := NewTestBlockchain("./test_db/testdb")
	assert.NoError(t, err)
	t.Cleanup(func() {
		// 测试结束后清理数据库
		bc.DB.Close()
	})
	return bc
}

// 测试合约创建
func TestContractCreation(t *testing.T) {
	bc := setupTestBlockchain(t)

	// 简单合约字节码 - 空合约
	code := []byte{0x60, 0x60, 0x60, 0x40, 0x52} // PUSH1 0x60 PUSH1 0x60 PUSH1 0x40 MSTORE

	// 创建合约
	account, err := bc.NewContract("0x6769811e7d937e8f849201f2e0e34e019db5f3bc", 1000, code)
	assert.NoError(t, err)
	assert.NotEmpty(t, account.State.Address)

	// 验证合约状态
	addr := common.HexToAddress(account.State.Address)
	assert.True(t, bc.StateDB.Exist(addr))
	assert.Equal(t, uint64(1), bc.StateDB.GetNonce(addr))
	assert.Equal(t, code, bc.StateDB.GetCode(addr))
}

// 测试简单合约调用 - 返回固定值
func TestSimpleContractCall(t *testing.T) {
	bc := setupTestBlockchain(t)

	// 部署一个简单合约 (返回42)
	codeHex := "6080604052348015600f57600080fd5b506004361060285760003560e01c80636d4ce63c14602d575b600080fd5b60336047565b604051603e91906067565b60405180910390f35b6000602a905090565b6000819050919050565b6061816050565b82525050565b6000602082019050607a6000830184605a565b9291505056fea2646970667358221220857e54f745071af59d359e42dfacba868b7c6695cb"
	code, err := hex.DecodeString(codeHex)
	assert.NoError(t, err, "Failed to decode bytecode")
	contract, err := bc.NewContract("0x6769811e7d937e8f849201f2e0e34e019db5f3bc", 1000, code)
	acc, err := bc.NewAccount()
	assert.NoError(t, err)

	// 合约ABI
	abiJSON := `[{"inputs":[],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`

	// 调用get()函数
	result, gasUsed, err := bc.RunCode(
		acc.State.Address, // 任意发送者地址
		contract.State.Address,
		abiJSON,
		1000000,
		"get",
	)
	assert.NoError(t, err)
	t.Logf("Gas used: %d", gasUsed)

	// 验证返回值
	assert.Equal(t, big.NewInt(42), result)
}

// 测试存储合约
func TestStorageContract(t *testing.T) {
	bc := setupTestBlockchain(t)

	// 从 solc 编译器获得的 runtime 字节码（仅部署后部分）
	bytecode := "608060405234801561001057600080fd5b50600436106100365760003560e01c80632e64cec11461003b5780636057361d14610059575b600080fd5b610043610075565b60405161005091906100a1565b60405180910390f35b610073600480360381019061006e91906100ed565b61007e565b005b60008054905090565b8060008190555050565b6000819050919050565b61009b81610088565b82525050565b60006020820190506100b66000830184610092565b92915050565b600080fd5b6100ca81610088565b81146100d557600080fd5b50565b6000813590506100e7816100c1565b92915050565b600060208284031215610103576101026100bc565b5b6000610111848285016100d8565b9150509291505056fea26469706673582212207bbff281ae8c9b6ea56e4da77beab4556e6bc57e474bf6d5e5b2e37f"
	code, err := hex.DecodeString(bytecode)
	assert.NoError(t, err, "Failed to decode bytecode")

	// 创建合约账户
	contract, err := bc.NewContract("0x6769811e7d937e8f849201f2e0e34e019db5f3bc", 1000, code)
	assert.NoError(t, err)

	// 合约ABI
	storageABI := `[{
		"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],
		"name":"store",
		"outputs":[],
		"stateMutability":"nonpayable",
		"type":"function"
	},{
		"inputs":[],
		"name":"retrieve",
		"outputs":[{"internalType":"uint256","name":"","type":"uint256"}],
		"stateMutability":"view",
		"type":"function"
	}]`

	// 创建发送者账户
	sender, err := bc.NewAccount()
	assert.NoError(t, err)

	t.Run("Store and Retrieve", func(t *testing.T) {
		// 调用 store(100)
		_, _, err := bc.RunCode(
			sender.State.Address,
			contract.State.Address,
			storageABI,
			1000000,
			"store",
			big.NewInt(100),
		)
		assert.NoError(t, err, "Call to store failed")

		// 调用 retrieve()
		result, _, err := bc.RunCode(
			sender.State.Address,
			contract.State.Address,
			storageABI,
			1000000,
			"retrieve",
		)
		assert.NoError(t, err, "Call to retrieve failed")

		assert.Equal(t, big.NewInt(100), result, "retrieve() should return 100")
	})
}

func TestRunEVMCall(t *testing.T) {
	bc := setupTestBlockchain(t)

	// 1. 合约 bytecode（仅 runtime 部分）
	bytecode := "608060405234801561001057600080fd5b50600436106100365760003560e01c80632e64cec11461003b5780636057361d14610059575b600080fd5b610043610075565b60405161005091906100a1565b60405180910390f35b610073600480360381019061006e91906100ed565b61007e565b005b60008054905090565b8060008190555050565b6000819050919050565b61009b81610088565b82525050565b60006020820190506100b66000830184610092565b92915050565b600080fd5b6100ca81610088565b81146100d557600080fd5b50565b6000813590506100e7816100c1565b92915050565b600060208284031215610103576101026100bc565b5b6000610111848285016100d8565b9150509291505056fea26469706673582212207bbff281ae8c9b6ea56e4da77beab4556e6bc57e474bf6d5e5b2e37f"

	code, err := hex.DecodeString(bytecode)
	assert.NoError(t, err, "Failed to decode bytecode")

	// 2. 创建合约账户
	contract, err := bc.NewContract("0x6769811e7d937e8f849201f2e0e34e019db5f3bc", 1000, code)
	assert.NoError(t, err)

	// 3. 创建发送者账户
	sender, err := bc.NewAccount()
	assert.NoError(t, err)

	// 4. 构造 store(100) 的 ABI 编码数据
	// 方法签名：store(uint256) → keccak256 前4字节 + padded 参数
	// 方法ID: keccak("store(uint256)")[:4] = 0x6057361d
	// 参数: 100 (uint256，填充32字节) = 0x...00000064
	storeData := "0x6057361d" + strings.Repeat("0", 64-len("64")) + "64"
	storeInput, err := hex.DecodeString(strings.TrimPrefix(storeData, "0x"))
	assert.NoError(t, err, "Invalid store input")

	// 5. 调用 RunEVMCall 执行 store(100)
	_, _, err = bc.RunEVMCall(
		sender.State.Address,
		contract.State.Address,
		storeInput,
		1000000,
		big.NewInt(0),
	)
	assert.NoError(t, err, "RunEVMCall store(100) failed")

	// 6. 构造 retrieve() 的调用数据（无参数）
	// 方法ID: keccak("retrieve()")[:4] = 0x2e64cec1
	retrieveInput, _ := hex.DecodeString("2e64cec1")

	// 7. 调用 RunEVMCall 执行 retrieve()
	output, _, err := bc.RunEVMCall(
		sender.State.Address,
		contract.State.Address,
		retrieveInput,
		1000000,
		big.NewInt(0),
	)
	assert.NoError(t, err, "RunEVMCall retrieve() failed")

	// 8. 解析返回的 output（应为 uint256 100）
	assert.Len(t, output, 32, "Output should be 32 bytes")
	result := new(big.Int).SetBytes(output)
	assert.Equal(t, big.NewInt(100), result, "retrieve() should return 100")
}
