

## LevelDB
```go
    //创建数据库
    levelDBPath := "./storage/leveldb_data"
	db, err := storage.NewLevelDB(levelDBPath)
	if err != nil {
		fmt.Println("Failed to create leveldb")
	}

	// 存储数据
	err = db.Put("nodeID", 7)
	if err != nil {
		fmt.Println("Failed to Put data")
	}

	// 读取数据
	var nodeID int
	err = db.Get("nodeID", &nodeID)
	if err != nil {
		fmt.Println("Failed to Get data")
	} else {
		fmt.Println(nodeID)
	}

    //删除数据
	err = db.Delete("nodeID")
	if err != nil {
		fmt.Println("Failed to Delete data")
	}
	defer db.Close()
    
```