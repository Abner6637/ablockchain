package core

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// 合约账户
func (bc *Blockchain) NewContract(address string, nonce uint64, balance uint64, code []byte) (*Account, error) {

	runtimeCode, _, _, err := bc.RunEVMCreate(address, code, 100000, big.NewInt(0))
	if err != nil {
		fmt.Println("bc.RunEVMCreate失败: " + err.Error())
	}
	contractAddr := CalculateContractAddress(address, nonce)

	accstate := &AccState{
		Address: contractAddr,
		Balance: balance,
		Nonce:   1,
		Code:    runtimeCode,
	}
	account := &Account{
		PrivateKey: nil,
		State:      accstate,
	}

	//暂存到内存中，不提交（产生交易后才提交）
	addr := common.HexToAddress(contractAddr)
	// if bc.StateDB.Exist(addr) {
	// 	return nil, fmt.Errorf("Account should not exist initially")
	// }

	bc.StateDB.CreateAccount(addr)
	if !bc.StateDB.Exist(addr) {
		return nil, fmt.Errorf("Account should exist after creation")
	}
	if !bc.StateDB.Empty(addr) {
		return nil, fmt.Errorf("New account should be empty")
	}
	// 创建后立即设置代码和初始状态
	bc.StateDB.SetCode(addr, runtimeCode)
	bc.StateDB.SetNonce(addr, 1, 0) // 合约账户初始nonce为1
	bc.StateDB.SetBalance(addr, uint256.NewInt(balance), 0)

	// 标记为合约账户
	bc.StateDB.SetState(addr, common.Hash{}, common.BytesToHash([]byte{1}))

	//s.stateObjects[account.State.Address] = accstate

	log.Printf("创建合约成功，地址: %s\n", contractAddr)
	bc.AddContrList(contractAddr)
	// 保存到本地
	_ = os.MkdirAll(privkeyaddr, 0700)
	err = SaveAccountToFile(account, privkeyaddr)
	if err != nil {
		return nil, fmt.Errorf("保存账户文件失败: %v", err)
	}

	return account, nil
}

func (bc *Blockchain) AddContrList(contract string) {
	// 检查账户是否已存在
	for _, acc := range bc.Contrlist {
		if acc == contract {
			return // 账户已存在，直接返回
		}
	}
	// 账户不存在，添加到列表
	bc.Contrlist = append(bc.Contrlist, contract)
}

// 收到部署合约的交易后，计算合约地址
//
//	func CalculateContractAddress(sender_address string, sender_nonce uint64) string {
//		//转换 address
//		address := common.HexToAddress(sender_address)
//		//构造 RLP 编码对象（数组 [address, nonce]）
//		rlpInput := []interface{}{address, sender_nonce}
//		//RLP 编码
//		encoded, err := rlp.EncodeToBytes(rlpInput)
//		if err != nil {
//			log.Fatalf("RLP encode failed: %v", err)
//		}
//		contract_address := gethcrypto.Keccak256(encoded)[12:]
//		fmt.Printf("计算合约地址: 0x%x\n", contract_address)
//		return fmt.Sprintf("0x%x", contract_address)
//	}
func CalculateContractAddress(sender string, nonce uint64) string {
	addr := gethcrypto.CreateAddress(common.HexToAddress(sender), nonce)
	return addr.Hex()
}

// 需要abi，暂时弃用
func (bc *Blockchain) RunCode(sender_addr string, contr_addr string, abi_string string, gas_limit uint64, function string, args ...interface{}) (interface{}, uint64, error) {
	// 参数检查
	if sender_addr == "" || contr_addr == "" {
		return nil, 0, errors.New("地址不能为空")
	}

	// 获取账户状态
	_, ok := bc.GetAccState(sender_addr)
	if !ok {
		return nil, 0, fmt.Errorf("发送者账户不存在: %s", sender_addr)
	}

	_, ok = bc.GetAccState(contr_addr)
	if !ok {
		return nil, 0, fmt.Errorf("合约账户不存在: %s", contr_addr)
	}

	// 解析ABI
	storageABI, err := abi.JSON(strings.NewReader(abi_string))
	if err != nil {
		return nil, 0, fmt.Errorf("ABI解析失败: %v", err)
	}

	// 准备调用参数
	contractAddr := common.HexToAddress(contr_addr)
	sender := common.HexToAddress(sender_addr)

	// 构造输入数据
	input, err := storageABI.Pack(function, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("参数打包失败: %v", err)
	}

	// 准备运行时配置
	runtimeCfg := runtime.Config{
		Origin:      sender,
		State:       bc.StateDB,
		GasLimit:    gas_limit,
		BlockNumber: bc.CurBlockNum,
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

	// 执行调用
	output, gasUsed, err := runtime.Call(contractAddr, input, &runtimeCfg)
	if err != nil {
		return nil, gasUsed, fmt.Errorf("EVM执行失败: %v", err)
	}

	// 处理返回值
	method, err := storageABI.MethodById(input[:4])
	if err != nil {
		return nil, gasUsed, fmt.Errorf("获取方法定义失败: %v", err)
	}

	// 动态创建结果变量
	var result interface{}
	if len(method.Outputs) > 0 {
		switch method.Outputs[0].Type.T {
		case abi.IntTy, abi.UintTy:
			var val *big.Int
			result = &val
		case abi.StringTy:
			var val string
			result = &val
		case abi.BoolTy:
			var val bool
			result = &val
		case abi.AddressTy:
			var val common.Address
			result = &val
		case abi.BytesTy, abi.FixedBytesTy:
			var val []byte
			result = &val
		case abi.ArrayTy, abi.SliceTy:
			// 处理数组类型...
		case abi.TupleTy:
			// 处理结构体类型...
		}
	}

	if result != nil {
		err = storageABI.UnpackIntoInterface(result, function, output)
		if err != nil {
			return nil, gasUsed, fmt.Errorf("结果解包失败: %v", err)
		}
		// 解引用指针
		result = reflect.ValueOf(result).Elem().Interface()
	}
	printResult(result)

	return result, gasUsed, nil
}

func printResult(result interface{}) {
	if result == nil {
		fmt.Println("执行结果为 nil")
		return
	}

	v := reflect.ValueOf(result)

	switch val := result.(type) {
	case *big.Int:
		fmt.Println("执行结果为 *big.Int:", val.String())
	case string:
		fmt.Println("执行结果为 string:", val)
	case bool:
		fmt.Println("执行结果为 bool:", val)
	case common.Address:
		fmt.Println("执行结果为 Address:", val.Hex())
	case []byte:
		fmt.Printf("执行结果为 []byte: 0x%x\n", val)
	default:
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			fmt.Printf("执行结果为数组，共 %d 个元素:\n", v.Len())
			for i := 0; i < v.Len(); i++ {
				fmt.Printf("  [%d] => %v\n", i, v.Index(i).Interface())
			}
		case reflect.Struct:
			fmt.Println("执行结果为结构体:")
			for i := 0; i < v.NumField(); i++ {
				fmt.Printf("  %s = %v\n", v.Type().Field(i).Name, v.Field(i).Interface())
			}
		default:
			fmt.Printf("未知类型（%T）: %v\n", val, val)
		}
	}
}

func (bc *Blockchain) RunEVMCall(senderAddr string, contractAddr string, input []byte, gasLimit uint64, value *big.Int) ([]byte, uint64, error) {
	if senderAddr == "" || contractAddr == "" {
		return nil, 0, errors.New("地址不能为空")
	}
	// 获取账户状态
	_, ok := bc.GetAccState(senderAddr)
	if !ok {
		return nil, 0, fmt.Errorf("发送者账户不存在: %s", senderAddr)
	}
	_, ok = bc.GetAccState(contractAddr)
	if !ok {
		return nil, 0, fmt.Errorf("合约账户不存在: %s", contractAddr)
	}
	sender := common.HexToAddress(senderAddr)
	contract := common.HexToAddress(contractAddr)

	// 构造运行时环境
	runtimeCfg := runtime.Config{
		Origin:      sender,
		State:       bc.StateDB,
		GasLimit:    gasLimit,
		Value:       value,
		BlockNumber: bc.CurBlockNum,
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
	// EVM 执行
	output, gasUsed, err := runtime.Call(contract, input, &runtimeCfg)
	if err != nil {
		return nil, gasUsed, fmt.Errorf("EVM执行失败: %v", err)
	}
	return output, gasUsed, nil
}

func (bc *Blockchain) RunEVMCreate(
	senderAddr string,
	deployCode []byte,
	gasLimit uint64,
	value *big.Int,
) ([]byte, []byte, uint64, error) {
	if senderAddr == "" {
		return nil, nil, 0, errors.New("发送者地址不能为空")
	}

	_, ok := bc.GetAccState(senderAddr)
	if !ok {
		return nil, nil, 0, fmt.Errorf("发送者账户不存在: %s", senderAddr)
	}

	sender := common.HexToAddress(senderAddr)

	runtimeCfg := runtime.Config{
		Origin:      sender,
		State:       bc.StateDB,
		GasLimit:    gasLimit,
		Value:       value,
		BlockNumber: bc.CurBlockNum,
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

	code, contractAddr, gasUsed, err := runtime.Create(deployCode, &runtimeCfg)
	if err != nil {
		return nil, nil, gasUsed, fmt.Errorf("runtime.Create失败: %v", err)
	}

	return code, contractAddr.Bytes(), gasUsed, nil
}
