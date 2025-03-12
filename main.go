package main

import (
	"fmt"
	"log"
	"os"

	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

type NodeConfig struct {
	NodeID    int    `json:"NodeID"`
	ChainID   int    `json:"ChainID"`
	Consensus string `json:"Consensus"`
}

func main() {

	nodedata, err := os.ReadFile("NodeConfig.json")
	if err != nil {
		fmt.Println("读取配置文件失败:", err)
		return
	}

	// 解析 JSON 数据
	var nodeconfig NodeConfig
	err = json.Unmarshal(nodedata, &nodeconfig)
	if err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return
	}

	// 打印解析后的配置
	fmt.Printf("配置: %+v\n", nodeconfig)

	// 打开（或创建）数据库
	db, err := leveldb.OpenFile("nodedb", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close() // 确保程序结束时关闭数据库

	// 写入数据
	err = db.Put([]byte("nodeid"), []byte("1"), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("数据存储成功！")

	// 读取数据
	data, err := db.Get([]byte("nodeid"), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("读取数据：", string(data))

	// 删除数据
	err = db.Delete([]byte("nodeid"), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("数据删除成功！")
}
