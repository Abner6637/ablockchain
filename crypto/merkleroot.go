package crypto

// 计算两个哈希的合并哈希
func HashPair(left, right []byte) []byte {
	hash, err := NewSHA256()
	if err != nil {
		return nil
	}

	data := append(left, right...)
	return hash.Hash(data)
}

// 计算默克尔树根
func ComputeMerkleRoot(txHashes [][]byte) []byte {
	if len(txHashes) == 0 {
		return nil
	}

	// 如果只有一个交易，直接返回它的哈希
	if len(txHashes) == 1 {
		return txHashes[0]
	}

	// 处理奇数情况：复制最后一个哈希
	if len(txHashes)%2 == 1 {
		txHashes = append(txHashes, txHashes[len(txHashes)-1])
	}

	// 计算下一层哈希
	var newLevel [][]byte
	for i := 0; i < len(txHashes); i += 2 {
		newHash := HashPair(txHashes[i], txHashes[i+1])
		newLevel = append(newLevel, newHash)
	}

	// 递归计算直到只剩一个哈希
	return ComputeMerkleRoot(newLevel)
}
