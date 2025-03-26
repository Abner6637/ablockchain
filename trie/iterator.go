package trie

// 迭代器,用于遍历Trie
type TrieIterator struct {
	nodes []*TrieNode
	keys  []string
	index int
}

func (t *Trie) NewIterator() *TrieIterator {
	var nodes []*TrieNode
	var keys []string
	collectNodes(t.Root, "", &nodes, &keys)

	return &TrieIterator{
		nodes: nodes,
		keys:  keys,
		index: 0,
	}
}

// 递归收集 Trie 中的所有账户
func collectNodes(node *TrieNode, prefix string, nodes *[]*TrieNode, keys *[]string) {
	if node == nil {
		return
	}

	// 如果当前节点存储了值，记录下来
	if node.Value != nil {
		*nodes = append(*nodes, node)
		*keys = append(*keys, prefix)
	}

	// 递归遍历子节点
	for char, child := range node.Children {
		collectNodes(child, prefix+string(char), nodes, keys)
	}
}

// 获取下一个账户
func (iter *TrieIterator) Next() bool {
	if iter.index < len(iter.nodes) {
		iter.index++
		return true
	}
	return false
}

// 获取当前账户的 Key
func (iter *TrieIterator) Key() string {
	if iter.index == 0 || iter.index > len(iter.keys) {
		return ""
	}
	return iter.keys[iter.index-1]
}

// 获取当前账户的 Value
func (iter *TrieIterator) Value() []byte {
	if iter.index == 0 || iter.index > len(iter.nodes) {
		return nil
	}
	return iter.nodes[iter.index-1].Value
}
