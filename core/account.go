package core

import (
	"ablockchain/crypto"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/rlp"
)

// 关于Account的操作在statedb文件中

type Account struct {
	PublicKey  []byte //由于*ecdsa.PublicKey不支持rlp编码,必须使用[]byte
	PrivateKey []byte

	State *AccState
}

// 记录账户状态，用于保存到MPT
type AccState struct {
	Address string
	Balance uint64
	Nonce   uint64
	Code    []byte // only for smart contract account
}

// 打包 签名和交易
type SignedTx struct {
	Tx   *Transaction
	Sign []byte
}

// 用于存储本地账户密钥的结构
type AccountKeyFile struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"` // hex or base64 encoded
}

func SaveAccountToFile(account *Account, dir string) error {
	keyFile := AccountKeyFile{
		Address:    account.State.Address,
		PrivateKey: hex.EncodeToString(account.PrivateKey),
	}

	data, err := json.MarshalIndent(keyFile, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化账户失败: %v", err)
	}

	// 文件名使用地址
	fileName := filepath.Join(dir, fmt.Sprintf("%s.json", account.State.Address))
	err = os.WriteFile(fileName, data, 0600)
	if err != nil {
		return fmt.Errorf("写入账户文件失败: %v", err)
	}
	return nil
}

func loadAccountKeysFromFile(address, dir string) ([]byte, error) {
	file := filepath.Join(dir, fmt.Sprintf("%s.json", address))
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("读取账户文件失败: %v", err)
	}

	var keyFile AccountKeyFile
	err = json.Unmarshal(data, &keyFile)
	if err != nil {
		return nil, fmt.Errorf("解析账户文件失败: %v", err)
	}

	privKey, err := hex.DecodeString(keyFile.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("解码私钥失败: %v", err)
	}

	return privKey, nil
}

// 对生成的交易进行签名, 返回rlp编码的SignedTx结构体
func SignTx(tx *Transaction) (*SignedTx, error) {
	privkey, err := loadAccountKeysFromFile(tx.From, privkeyaddr)
	if err != nil {
		fmt.Println("没有该账户私钥，无法签名交易") //待docker测试
		return nil, err
	}
	encodetx, err := tx.EncodeTx()
	if err != nil {
		fmt.Println("交易编码失败")
		return nil, err
	}
	hashTx := crypto.GlobalHashAlgorithm.Hash(encodetx) //编码后计算hash
	privk, err := crypto.ToECDSA(privkey)
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
