#!/bin/bash

# 搭建三个节点的区块链系统

# 删除旧的 block_storage 文件
rm -rf ./block_storage
mkdir -p ./block_storage/

# 设置数据库目录
NODE1_DB="./block_storage/node1_data"
NODE2_DB="./block_storage/node2_data"
NODE3_DB="./block_storage/node3_data"

NODE1_LISTEN="/ip4/127.0.0.1/tcp/50001"
NODE2_LISTEN="/ip4/127.0.0.1/tcp/50002"
NODE3_LISTEN="/ip4/127.0.0.1/tcp/50003"

NODE1_LOG="./block_storage/node1.log"
NODE2_LOG="./block_storage/node2.log"
NODE3_LOG="./block_storage/node3.log"


# 启动第一个节点，并将其输出重定向到 log 文件
tmux new-session -d -s node1 "go run main.go --db $NODE1_DB --listen $NODE1_LISTEN --consensus pow | tee $NODE1_LOG"

# 等待一会儿，确保节点1启动完成
sleep 5

# 解析节点1的 p2p 地址
NODE1_PEER_ID=$(grep -oP 'ID:\s+\K[a-zA-Z0-9]+' $NODE1_LOG | tail -1)

NODE1_ADDR="$NODE1_LISTEN/p2p/$NODE1_PEER_ID"

echo "Node 1 Address: $NODE1_ADDR"

# 启动第二个节点
tmux new-session -d -s node2 "go run main.go --db $NODE2_DB --listen $NODE2_LISTEN --consensus pow | tee $NODE2_LOG"

sleep 5

NODE2_PEER_ID=$(grep -oP 'ID:\s+\K[a-zA-Z0-9]+' $NODE2_LOG | tail -1)

NODE2_ADDR="$NODE2_LISTEN/p2p/$NODE2_PEER_ID"

echo "Node 2 Address: $NODE2_ADDR"

# 启动第三个节点
tmux new-session -d -s node3 "go run main.go --db $NODE3_DB --listen $NODE3_LISTEN --consensus pow | tee $NODE3_LOG"

sleep 5

NODE3_PEER_ID=$(grep -oP 'ID:\s+\K[a-zA-Z0-9]+' $NODE3_LOG | tail -1)

NODE3_ADDR="$NODE3_LISTEN/p2p/$NODE3_PEER_ID"

echo "Node 3 Address: $NODE3_ADDR"

# 节点连接
tmux send-keys -t node1 "connect $NODE2_ADDR" Enter
tmux send-keys -t node2 "connect $NODE3_ADDR" Enter
tmux send-keys -t node3 "connect $NODE1_ADDR" Enter

echo "Blockchain Start!"

