package main

import (
	"ablockchain/cli"
	"ablockchain/event"
	"ablockchain/system"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func attachToSocket() {
	conn, err := net.Dial("unix", "/tmp/ablockchain.sock")
	if err != nil {
		log.Fatalf("连接主进程失败: %v", err)
	}
	defer conn.Close()

	fmt.Println("已连接到 ablockchain 节点。输入命令，Ctrl+C 退出。")

	go io.Copy(os.Stdout, conn)
	io.Copy(conn, os.Stdin)
}

func main() {
	// log默认输出到stderr，将其重定向至stdout，当使用tee时可显示
	// tee默认读取stdout流
	// 或者使用tee时，前面加上”2>&1“
	//log.SetOutput(os.Stdout)

	// 解析命令行参数
	cfg, isAttach := cli.ParseFlags()

	exitEvent := event.Bus.Subscribe("ExitEvent")

	// log.Printf("is attach: %v", isAttach)

	if !isAttach {
		system.StartSystem(cfg)
		select {
		case msg := <-exitEvent:
			exit, _ := msg.(bool)
			log.Printf("接收到exit命令，退出进程")
			if exit {
				return
			}
		}
	} else {
		fmt.Println("正在打开交互界面...")
		attachToSocket()
	}
	// http.HandleFunc("/", rpc.Handler)

	// fmt.Println("RPC Server listening on http://localhost:8545")
	// if err := http.ListenAndServe(":8545", nil); err != nil {
	// 	log.Fatalf("启动失败: %v", err)
	// }
}
