package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
)

// ContractExecutionEnv 智能合约执行环境
type ContractExecutionEnv struct {
	ChainConfig *params.ChainConfig // 区块链配置
	GasLimit    uint64              // Gas限制
}

// NewContractExecutionEnv 创建新的合约执行环境
func NewContractExecutionEnv() *ContractExecutionEnv {
	return &ContractExecutionEnv{
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      big.NewInt(0),
			DAOForkBlock:        nil,
			DAOForkSupport:      true,
			EIP150Block:         big.NewInt(0),
			EIP155Block:         big.NewInt(0),
			EIP158Block:         big.NewInt(0),
			ByzantiumBlock:      big.NewInt(0),
			ConstantinopleBlock: big.NewInt(0),
			PetersburgBlock:     big.NewInt(0),
			IstanbulBlock:       big.NewInt(0),
			MuirGlacierBlock:    big.NewInt(0),
			BerlinBlock:         big.NewInt(0),
			LondonBlock:         big.NewInt(0),
		},
		GasLimit: 10000000, // 默认Gas限制
	}
}

// ExecuteContract 执行智能合约
func (env *ContractExecutionEnv) ExecuteContract(
	caller common.Address, // 调用者地址
	contractAddr common.Address, // 合约地址
	contractCode []byte, // 合约字节码
	input []byte, // 调用输入数据
	value *big.Int, // 转账金额
	stateDB *state.StateDB, // 状态数据库
) ([]byte, uint64, error) {

	// 准备运行时配置
	config := runtime.Config{
		Origin:      caller,
		State:       stateDB,
		GasLimit:    env.GasLimit,
		GasPrice:    big.NewInt(0),
		Value:       value,
		Difficulty:  big.NewInt(1),
		Time:        uint64(0),
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(1),
		ChainConfig: env.ChainConfig,
		EVMConfig:   vm.Config{},
	}

	// 如果合约地址不存在，先创建合约账户
	if !stateDB.Exist(contractAddr) {
		stateDB.CreateAccount(contractAddr)
	}

	// 设置合约代码
	stateDB.SetCode(contractAddr, contractCode)

	// 执行合约调用
	return runtime.Call(contractAddr, input, &config)
}

// 示例使用
//func main() {
// // 1. 创建执行环境
// env := NewContractExecutionEnv()

// // 2. 准备一个简单的状态数据库实现
// // 这里使用一个内存实现，实际项目中应该用持久化存储
// stateDB := NewMemoryStateDB()

// // 3. 定义调用者和合约地址
// caller := common.HexToAddress("0xCaller")
// contractAddr := common.HexToAddress("0xContract")

// // 4. 准备一个简单的合约字节码
// // 这个合约只是返回输入数据的长度
// contractCode := common.FromHex("0x6080604052348015600f57600080fd5b506004361060285760003560e01c8063a5af1f1d14602d575b600080fd5b60336047565b604051603e9190605a565b60405180910390f35b60006001905090565b6054816073565b82525050565b6000602082019050606d6000830184604d565b92915050565b600081905091905056")

// // 5. 准备调用输入数据
// input := common.FromHex("0xa5af1f1d") // 函数选择器

// alloc := make(types.GenesisAlloc)
// alloc[1] = types.Account{
// 	Nonce:   0,
// 	Code:    a.Bytes(),
// 	Balance: big.NewInt(0xffffffff),
// }
// alloc[2] = types.Account{
// 	Nonce:   0,
// 	Code:    b.Bytes(),
// 	Balance: big.NewInt(0xffffffff),
// }
// var (
// 	statedb = StateDBWithAlloc(alloc)
// 	sender  = common.BytesToAddress([]byte("sender"))
// )

// // 6. 执行合约
// ret, remainingGas, err := env.ExecuteContract(
// 	caller,
// 	contractAddr,
// 	contractCode,
// 	input,
// 	big.NewInt(0), // 不转账
// 	stateDB,
// )

// // 7. 输出结果
// fmt.Printf("Execution results:\n")
// fmt.Printf("Return data: 0x%x\n", ret)
// fmt.Printf("Remaining gas: %d\n", remainingGas)
// fmt.Printf("Error: %v\n", err)

// // 8. 打印合约账户状态
// code := stateDB.GetCode(contractAddr)
// fmt.Printf("Contract code size: %d bytes\n", len(code))
//}
