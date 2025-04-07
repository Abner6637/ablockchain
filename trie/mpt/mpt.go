package mpt

import (
	"ablockchain/crypto"
	"ablockchain/storage"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/rlp"
)

type NodeType int

const (
	LeafNode NodeType = iota
	ExtensionNode
	BranchNode
)

type Node struct {
	Type     NodeType
	Path     string
	Value    []byte
	Children []*Node // 对于 Branch 节点有 16 个分支
	Hash     []byte
}

type Trie struct {
	Root *Node
	DB   *storage.LevelDB
}

func NewTrie(dbPath string) (*Trie, error) {
	db, err := storage.NewLevelDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &Trie{
		Root: &Node{
			Type:     BranchNode,
			Children: make([]*Node, 16),
		},
		DB: db,
	}, nil
}

// 将账户数据插入到 Trie 中，同时将地址和账户数据也存入 LevelDB
func (t *Trie) Insert(key string, value []byte) error {
	node := t.Root
	for i := 0; i < len(key); i++ {
		index := int(key[i]) % 16
		if node.Children[index] == nil {
			node.Children[index] = &Node{
				Type:     BranchNode,
				Children: make([]*Node, 16),
			}
		}
		node = node.Children[index]
	}
	node.Type = LeafNode
	node.Path = key
	node.Value = value
	if t.DB != nil {
		return t.DB.Put(key, value)
	}
	return nil
}

// 根据地址获取账户数据
func (t *Trie) Get(key string) ([]byte, error) {
	node := t.Root
	for i := 0; i < len(key); i++ {
		index := int(key[i]) % 16
		if node.Children[index] == nil {
			return nil, errors.New("账户不存在")
		}
		node = node.Children[index]
	}
	if node.Type == LeafNode {
		return node.Value, nil
	}
	return nil, errors.New("数据为空")
}

// 将整棵 Trie 的所有节点哈希计算，并返回根哈希，同时把节点 RLP 编码写入 DB（hash -> node）
func (t *Trie) Commit() ([]byte, error) {
	if t.Root == nil {
		return nil, nil
	}
	return t.commitNode(t.Root)
}

func (t *Trie) commitNode(node *Node) ([]byte, error) {
	var rlpEncoded []byte
	var err error

	switch node.Type {
	case LeafNode:
		rlpEncoded, err = rlp.EncodeToBytes([]interface{}{node.Path, node.Value})
	case ExtensionNode:
		childHash, _ := t.commitNode(node.Children[0])
		rlpEncoded, err = rlp.EncodeToBytes([]interface{}{node.Path, childHash})
	case BranchNode:
		var children [16][]byte
		for i := 0; i < 16; i++ {
			if node.Children[i] != nil {
				children[i], _ = t.commitNode(node.Children[i])
			} else {
				children[i] = []byte{}
			}
		}
		list := make([]interface{}, 17)
		for i := 0; i < 16; i++ {
			list[i] = children[i]
		}
		list[16] = node.Value
		rlpEncoded, err = rlp.EncodeToBytes(list)
	}

	if err != nil {
		return nil, err
	}

	hash := crypto.GlobalHashAlgorithm.Hash(rlpEncoded)
	node.Hash = hash
	hashstr := hex.EncodeToString(hash)

	if t.DB != nil {
		t.DB.Put(hashstr, rlpEncoded)
	}
	return hash, nil
}
