#!/bin/bash

read -p "请输入节点数量: " NODE_COUNT
START_IP=10

# 1. 生成地址文件
go run keygen/main.go keygen -n ${NODE_COUNT}

# 2. 拼接 valaddress 参数
VALADDR_ARGS=""
for ((i=0; i<NODE_COUNT; i++)); do
  ADDRESS=$(cat ./key_store/address${i})
  VALADDR_ARGS+="--valaddress ${ADDRESS} "
done

PEERADDR_ARGS=""
for ((i=0; i<NODE_COUNT; i++)); do
  PEERADDR=$(cat ./key_store/p2pID${i})
  PEERADDR_ARGS+="--peer /ip4/172.28.0.$((START_IP + i))/tcp/32000/p2p/${PEERADDR} "
done

# 3. 写入 docker-compose.yml 头部
cat > docker-compose.yml <<EOF
version: '3.9'

services:
EOF

# 4. 写入每个节点的服务配置
for ((i=0; i<NODE_COUNT; i++)); do
  node_ip=$((START_IP + i))
  cat >> docker-compose.yml <<EOF
  node${i}:
    container_name: node${i}
    depends_on:
      - bootnode
    image: ablockchain:0.0.5
    networks:
      customnet:
        ipv4_address: 172.28.0.${node_ip}
    ports:
      - "3300${i}:32000"
    volumes:
      - ./key_store/nodekey${i}:/ablockchain/key_store/nodekey${i}
      - ./genesis.json:/ablockchain/genesis.json
    command: >
      ./ablockchain 
      --listen /ip4/0.0.0.0/tcp/32000 
      --nodekey ./key_store/nodekey${i} 
      --consensus hotstuff
      ${VALADDR_ARGS}
      ${PEERADDR_ARGS}
      --backend=true
EOF
done

# 5. 写入网络配置
cat >> docker-compose.yml <<EOF
networks:
  customnet:
    driver: bridge
    ipam:
      config:
        - subnet: 172.28.0.0/16
EOF

echo "✅ docker-compose.yml 生成完毕！请使用 docker-compose up -d 启动集群。"
