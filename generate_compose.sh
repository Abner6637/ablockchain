# #!/bin/bash

# # 目前有个问题，最多只能支持10个节点
# read -p "请输入节点数量: " NODE_COUNT
# START_IP=10

# # 1. 生成地址文件
# go run keygen/main.go keygen -n ${NODE_COUNT}

# # 2. 拼接 valaddress 参数
# VALADDR_ARGS=""
# for ((i=0; i<NODE_COUNT; i++)); do
#   ADDRESS=$(cat ./key_store/address${i})
#   VALADDR_ARGS+="--valaddress ${ADDRESS} "
# done

# PEERADDR_ARGS=""
# for ((i=0; i<NODE_COUNT; i++)); do
#   PEERADDR=$(cat ./key_store/p2pID${i})
#   PEERADDR_ARGS+="--peer /ip4/172.29.0.$((START_IP + i))/tcp/32000/p2p/${PEERADDR} "
# done

# # 3. 写入 docker-compose.yml 头部
# cat > docker-compose.yml <<EOF
# version: '3.9'

# services:
# EOF

# # 4. 写入每个节点的服务配置
# for ((i=0; i<NODE_COUNT; i++)); do
#   node_ip=$((START_IP + i))
#   cat >> docker-compose.yml <<EOF
#   node$((i+10)):
#     container_name: node$((i+10))
#     depends_on:
#       - bootnode
#     image: ablockchain:6.9.3
#     networks:
#       customnet1:
#         ipv4_address: 172.29.0.${node_ip}
#     ports:
#       - "3310${i}:32000"
#       - "2610${i}:8545"
#       - "2210${i}:5173"
#     volumes:
#       - ./key_store/nodekey${i}:/ablockchain/key_store/nodekey${i}
#       - ./genesis.json:/ablockchain/genesis.json
#     environment:
#       - API_URL=http://192.168.2.127:2610${i}
#     command: >
#       ./ablockchain 
#       --listen /ip4/0.0.0.0/tcp/32000 
#       --nodekey ./key_store/nodekey${i} 
#       --consensus poa
#       ${VALADDR_ARGS}
#       ${PEERADDR_ARGS}
#       --backend=true
# EOF
# done

# # 5. 写入网络配置
# cat >> docker-compose.yml <<EOF
# networks:
#   customnet1:
#     driver: bridge
#     ipam:
#       config:
#         - subnet: 172.29.0.0/16
# EOF

# echo "✅ docker-compose.yml 生成完毕！请使用 docker-compose up -d 启动集群。"
#!/bin/bash

read -p "请输入节点数量: " NODE_COUNT
START_IP=10

# 1. 生成地址文件
go run keygen/main.go keygen -n ${NODE_COUNT}

# 2. 拼接验证者地址参数（所有节点共享）
VALADDR_ARGS=""
for ((i=0; i<NODE_COUNT; i++)); do
  ADDRESS=$(cat ./key_store/address${i})
  VALADDR_ARGS+="--valaddress ${ADDRESS} "
done

# 3. 写入 docker-compose.yml 头部
cat > docker-compose.yml <<EOF
version: '3.9'

services:
EOF

# 4. 写入每个节点的服务配置（每个节点只连接其他节点）
for ((i=0; i<NODE_COUNT; i++)); do
  node_ip=$((START_IP + i))

  # 构建当前节点的 peer 参数
  PEER_ARGS=""
  for ((j=0; j<NODE_COUNT; j++)); do
    if [ "$j" -ne "$i" ]; then
      peer_ip=$((START_IP + j))
      peer_id=$(cat ./key_store/p2pID${j})
      PEER_ARGS+="--peer /ip4/172.29.0.${peer_ip}/tcp/32000/p2p/${peer_id} "
    fi
  done

  cat >> docker-compose.yml <<EOF
  node$((i+10)):
    container_name: node$((i+10))
    depends_on:
      - bootnode
    image: ablockchain:9.9.9
    networks:
      customnet1:
        ipv4_address: 172.29.0.${node_ip}
    ports:
      - "3310${i}:32000"
      - "2610${i}:8545"
      - "2210${i}:5173"
    volumes:
      - ./key_store/nodekey${i}:/ablockchain/key_store/nodekey${i}
      - ./genesis.json:/ablockchain/genesis.json
    environment:
      - API_URL=http://192.168.2.127:2610${i}
    command: >
      ./ablockchain 
      --listen /ip4/0.0.0.0/tcp/32000 
      --nodekey ./key_store/nodekey${i} 
      --consensus poa 
      ${VALADDR_ARGS}
      ${PEER_ARGS}
      --backend=true
EOF
done

# 5. 写入网络配置
cat >> docker-compose.yml <<EOF

networks:
  customnet1:
    driver: bridge
    ipam:
      config:
        - subnet: 172.29.0.0/16
EOF

echo "✅ docker-compose.yml 生成完毕！请使用 docker-compose up -d 启动集群。"
