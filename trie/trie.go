package trie

import (
	"ablockchain/crypto"
	"ablockchain/storage"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/rlp"
)

type TrieNode struct {
	Children map[byte]*TrieNode // 子节点
	Value    []byte             // 存储账户数据
	Hash     string             // 当前节点的哈希值
}

type Trie struct {
	Root *TrieNode
	DB   *storage.LevelDB
}

func NewTrie(dbPath string) (*Trie, error) {
	db, err := storage.NewLevelDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &Trie{
		Root: &TrieNode{Children: make(map[byte]*TrieNode)},
		DB:   db,
	}, nil
}

// 计算节点哈希
func hashNode(node *TrieNode) string {
	data, _ := rlp.EncodeToBytes(node)
	hash := crypto.GlobalHashAlgorithm.Hash(data)
	return hex.EncodeToString(hash[:])
}

// 插入或更新账户信息
func (t *Trie) Insert(key string, value []byte) error {
	node := t.Root
	for i := 0; i < len(key); i++ {
		if node.Children[key[i]] == nil {
			node.Children[key[i]] = &TrieNode{Children: make(map[byte]*TrieNode)}
		}
		node = node.Children[key[i]]
	}
	node.Value = value
	node.Hash = hashNode(node)

	return t.DB.Put(key, value)
}

// 查找账户信息
func (t *Trie) Get(key string) ([]byte, error) {
	node := t.Root
	for i := 0; i < len(key); i++ {
		if node.Children[key[i]] == nil {
			return nil, errors.New("账户不存在")
		}
		node = node.Children[key[i]]
	}
	if node.Value == nil {
		return nil, errors.New("账户数据为空")
	}
	return node.Value, nil
}

// 计算整个 Trie 的根哈希
func (t *Trie) RootHash() []byte {
	if t.Root == nil {
		return nil
	}
	return t.computeNodeHash(t.Root)
}

// 递归计算 TrieNode 的哈希值
func (t *Trie) computeNodeHash(node *TrieNode) []byte {
	// 若为叶子节点，直接计算其哈希
	if len(node.Children) == 0 {
		if node.Value == nil {
			return nil
		}
		return crypto.GlobalHashAlgorithm.Hash(node.Value)
	}

	// 递归计算子节点哈希
	var childHashes [][]byte
	for key, child := range node.Children {
		childHash := t.computeNodeHash(child)
		if childHash != nil {
			childHashes = append(childHashes, append([]byte{key}, childHash...))
		}
	}

	// 编码子节点哈希列表和当前节点值
	data, _ := rlp.EncodeToBytes(struct {
		Value       []byte
		ChildHashes [][]byte
	}{
		Value:       node.Value,
		ChildHashes: childHashes,
	})

	// 计算当前节点哈希
	node.Hash = hex.EncodeToString(crypto.GlobalHashAlgorithm.Hash(data))
	return crypto.GlobalHashAlgorithm.Hash(data)
}
