package core

import (
	"ablockchain/crypto"
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
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// 合约账户
func (bc *Blockchain) NewContract(balance uint64, code []byte) (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	privKeyBytes := crypto.FromECDSA(privKey)
	accstate := &AccState{
		Address: address,
		Balance: balance,
		Nonce:   0,
		Code:    code,
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
	// 创建后立即设置代码和初始状态
	bc.StateDB.SetCode(addr, code)
	bc.StateDB.SetNonce(addr, 1, 0) // 合约账户初始nonce为1
	bc.StateDB.SetBalance(addr, uint256.NewInt(balance), 0)

	// 标记为合约账户
	bc.StateDB.SetState(addr, common.Hash{}, common.BytesToHash([]byte{1}))

	//s.stateObjects[account.State.Address] = accstate

	log.Printf("创建合约成功，地址: %s\n", address)
	bc.AddContrList(address)
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
