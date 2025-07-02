package core

import (
	"encoding/json"
	"fmt"
	"math/big"
)

/*在与remix的通信中,需要用交易hash查询交易数据.
为了实现这一功能,添加区块和交易的索引,存储在一个独立leveldb中*/

// 定义索引结构：TxIndex 表示一个交易在区块中对应的序号
type TxIndex struct {
	BlockHash []byte
	TxIndex   int
}

// 索引构建函数
func (bc *Blockchain) StoreTxIndex(block *Block) error {
	blockHash := block.Hash

	// 构建 tx:<txHash> → BlockHash + TxIndex
	for i, tx := range block.Transactions {
		// 构建 TxIndex 数据
		index := TxIndex{
			BlockHash: blockHash,
			TxIndex:   i,
		}
		indexBytes, _ := json.Marshal(index) //可替换为rlp编码

		txKey := append([]byte("tx:"), tx.TxHash...)
		if err := bc.IndexDB.Put(txKey, indexBytes, nil); err != nil {
			return err
		}

		// 地址索引 addr:<address>:<blockNum> → [txHash1, txHash2...]
		// addrKey := fmt.Sprintf("addr:%s:%s", tx.From, block.Header.Number.String())
		// prev, _ := bc.IndexDB.Get([]byte(addrKey), nil)
		// var txList [][]byte
		// if prev != nil {
		// 	json.Unmarshal(prev, &txList)
		// }
		// txList = append(txList, tx.TxHash)
		// newVal, _ := json.Marshal(txList)
		// bc.IndexDB.Put([]byte(addrKey), newVal, nil)
	}

	// 构建 block:<blockHash>:txs → TxHash 列表
	// var txHashes [][]byte
	// for _, tx := range block.Transactions {
	// 	txHashes = append(txHashes, tx.TxHash)
	// }
	// txListBytes, _ := json.Marshal(txHashes)
	// blockTxKey := append([]byte("block:"), append(blockHash, []byte(":txs")...)...)
	// if err := bc.IndexDB.Put(blockTxKey, txListBytes, nil); err != nil {
	// 	return err
	// }

	return nil
}

// 查询函数：根据 txHash 查询交易位置
func (bc *Blockchain) QueryTxIndex(txHash []byte) (*TxIndex, error) {
	txKey := append([]byte("tx:"), txHash...)
	data, err := bc.IndexDB.Get(txKey, nil)
	if err != nil {
		return nil, err
	}
	var index TxIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// 由TxIndex查询得到交易本体
func (bc *Blockchain) GetTxByIndex(index *TxIndex) *Transaction {
	block, err := bc.GetBlockByHash(index.BlockHash)
	if err != nil {
		return nil
	}
	return block.Transactions[index.TxIndex]
}

func (bc *Blockchain) GetTxByHash(txHash []byte) (*Transaction, error) {
	index, err := bc.QueryTxIndex(txHash)
	if err != nil {
		return nil, err
	}
	return bc.GetTxByIndex(index), nil
}

// ---------------------------------------------------------------------------------

// 保存区块哈希->区块高度的索引
func (bc *Blockchain) StoreBlockIndex(block *Block) error {
	// 编号转字节
	blockNumBytes := block.Header.Number.Bytes()
	// 保存哈希 → 编号的索引
	hashKey := append([]byte("hash2num:"), block.Hash...)
	if err := bc.IndexDB.Put(hashKey, blockNumBytes, nil); err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) GetBlockNumByHash(blockHash []byte) (string, error) {
	// 查找哈希 → 编号
	hashKey := append([]byte("hash2num:"), blockHash...)
	blockNumBytes, err := bc.IndexDB.Get(hashKey, nil)
	if err != nil {
		return "", fmt.Errorf("hash not found: %w", err)
	}
	// 转为 string
	blockNumberStr := new(big.Int).SetBytes(blockNumBytes).String()
	return blockNumberStr, nil
}

func (bc *Blockchain) GetBlockByNum(num string) (*Block, error) {
	// 查找编号 → 区块
	var block Block
	err := bc.DB.Get(num, &block)
	if err != nil {
		return nil, fmt.Errorf("读取区块 %s 失败: %w", num, err)
	}
	return &block, nil
}

func (bc *Blockchain) GetBlockByHash(blockHash []byte) (*Block, error) {
	// 查找哈希 → 编号
	hashKey := append([]byte("hash2num:"), blockHash...)
	blockNumBytes, err := bc.IndexDB.Get(hashKey, nil)
	if err != nil {
		return nil, fmt.Errorf("hash not found: %w", err)
	}
	// 转为 string
	blockNumberStr := new(big.Int).SetBytes(blockNumBytes).String()
	// 查找编号 → 区块
	var block Block
	err = bc.DB.Get(blockNumberStr, &block)
	if err != nil {
		return nil, fmt.Errorf("读取区块 %s 失败: %w", blockNumberStr, err)
	}
	return &block, nil
}
