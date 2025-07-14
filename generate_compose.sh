#!/bin/bash

# 目前有个问题，最多只能支持10个节点
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
  PEERADDR_ARGS+="--peer /ip4/172.30.0.$((START_IP + i))/tcp/32000/p2p/${PEERADDR} "
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
  node$((i + 20)):
    container_name: node$((i + 20))
    depends_on:
      - bootnode
    image: ablockchain:2.2.8
    networks:
      customnet2:
        ipv4_address: 172.30.0.${node_ip}
    ports:
      - "3320${i}:32000"
      - "2620${i}:8545"
      - "2220${i}:5173"
    volumes:
      - ./key_store/nodekey${i}:/ablockchain/key_store/nodekey${i}
      - ./genesis.json:/ablockchain/genesis.json
    environment:
      - API_URL=http://192.168.2.127:2620${i}
    command: >
      ./ablockchain 
      --listen /ip4/0.0.0.0/tcp/32000 
      --nodekey ./key_store/nodekey${i} 
      --consensus poa
      ${VALADDR_ARGS}
      ${PEERADDR_ARGS}
      --backend=true
EOF
done

# 5. 写入网络配置
cat >> docker-compose.yml <<EOF
networks:
  customnet2:
    driver: bridge
    ipam:
      config:
        - subnet: 172.30.0.0/16
EOF

echo "✅ docker-compose.yml 生成完毕！请使用 docker-compose up -d 启动集群。"
