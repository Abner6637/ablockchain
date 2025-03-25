#!/bin/bash

# 终端颜色
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# 删除旧的 multiaddr 文件
rm -rf ./block_storage
mkdir -p ./block_storage/

echo -e "${GREEN}启动第一个节点...${NC}"
go run main.go --db ./block_storage/node1_data --consensus pow
# go run main.go --db ./block_storage/node2_data --consensus pow

# 等待节点启动
sleep 3

# # 提取第一个节点的 Multiaddr
# NODE1_ADDR=$(grep -oE '/ip4[^ ]+' ./block_storage/node1_addr.txt | head -n 1)
# if [ -z "$NODE1_ADDR" ]; then
#   echo "未能获取 Node 1 的地址"
#   exit 1
# fi
# echo -e "${GREEN}Node 1 地址: ${NODE1_ADDR}${NC}"

# echo -e "${GREEN}启动第二个节点...${NC}"
# gnome-terminal -- bash -c "go run main.go --db ./block_storage/node2_data | tee ./block_storage/node2_addr.txt"

# # 等待第二个节点启动
# sleep 5

# echo -e "${GREEN}Node 2 连接到 Node 1...${NC}"
# echo "connect $NODE1_ADDR" | nc -w 1 localhost 9001

# echo -e "${GREEN}两个节点已成功启动并连接!${NC}"
