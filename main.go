package main

import (
	"ablockchain/storage"
	"fmt"
)

func main() {
	levelDBPath := "./storage/leveldb_data"
	db, err := storage.NewLevelDB(levelDBPath)
	if err != nil {
		fmt.Println("Failed1")
	}

	// 存储数据
	err = db.Put("nodeID", 001)
	if err != nil {
		fmt.Println("Failed2")
	}

	// 读取数据
	var nodeID int
	err = db.Get("nodeID", &nodeID)
	if err != nil {
		fmt.Println("Failed3")
	} else {
		fmt.Println(nodeID)
	}
	defer db.Close()
}
