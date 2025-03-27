#!/bin/bash

# 使用 bash mesh.sh 运行

# 设置节点数量
NODE_COUNT=4

# 删除旧的 block_storage 目录并重新创建
rm -rf ./block_storage
mkdir -p ./block_storage

rm -rf ./key_store
mkdir -p ./key_store

# 监听端口的起始值
BASE_PORT=32000

# 存储所有节点的 P2P 地址
declare -a NODE_ADDRS

# 启动所有节点
for ((i=1; i<=NODE_COUNT; i++)); do
    NODE_DB="./block_storage/node${i}_data"
    NODE_LOG="./block_storage/node${i}.log"
    NODE_PORT=$((BASE_PORT + i))
    NODE_LISTEN="/ip4/127.0.0.1/tcp/${NODE_PORT}"
    NODE_NODEKEY="./key_store/nodekey${i}"
    # NODE_CONCENSUSNUM=NODE_COUNT

    # 启动当前节点
    tmux new-session -d -s "node${i}" "go run main.go --db $NODE_DB --listen $NODE_LISTEN \
    --consensus pbft --nodekey $NODE_NODEKEY 2>&1 | tee $NODE_LOG"

    # 确保节点启动完成
    sleep 5

    # 解析节点 P2P ID
    NODE_PEER_ID=$(grep -oP 'ID:\s+\K[a-zA-Z0-9]+' $NODE_LOG | tail -1)
    NODE_ADDR="${NODE_LISTEN}/p2p/${NODE_PEER_ID}"

    echo "节点 $i 地址: $NODE_ADDR"

    # 存储该节点的地址
    NODE_ADDRS+=("$NODE_ADDR")
done

# 让所有节点相互连接（形成 Mesh 结构）
for ((i=1; i<=NODE_COUNT; i++)); do
    for ((j=1; j<=i; j++)); do
        if [[ $i -ne $j ]]; then  # 避免自己连接自己
            echo "节点 $i 连接到节点 $j..."
            tmux send-keys -t "node${i}" "connect ${NODE_ADDRS[$((j-1))]}" Enter
            sleep 1 
        fi
    done
done

echo "所有节点已启动并相互连接"
