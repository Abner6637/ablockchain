package main

import (
	"ablockchain/storage"
	"fmt"
)

func main() {
	dbPath := "test.db"
	bucketName := "TestBucket"
	db, err := storage.NewBoltDB(dbPath, bucketName)
	if err != nil {
		fmt.Println("Failed1")
	}

	// 存储数据
	err = db.Put("nodeID", 7)
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

	err = db.Delete("nodeID")
	if err != nil {
		fmt.Println("Failed to Delete data")
	}

	defer db.Close()
}
