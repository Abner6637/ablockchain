#!/bin/bash

read -p "请输入节点数量: " NODE_COUNT
START_IP=10

cat > docker-compose.yml <<EOF
version: '3.9'

x-node-defaults: &node-defaults
  image: ablockchain:0.0.0
  networks:
    customnet:
      ipv4_address: 172.28.0.\${NODE_IP}
  ports:
    - "3300\${NODE_ID}:32000"
  volumes:
    - ./key_store/nodekey\${NODE_ID}:/ablockchain/key_store/nodekey\${NODE_ID}
  command: >
    ./ablockchain 
    --listen /ip4/0.0.0.0/tcp/32000 
    --nodekey ./key_store/nodekey\${NODE_ID} 
    --valaddress 0x11820fda9ed706f0b98140e3b3b8b92208fac562 
    --valaddress 0xd870276849a24fac70f2509e5d352f7dd9ebb49b
    --valaddress 0x8dff8c16895f70fa9646b57932ae0d6c16427b05 
    --valaddress 0x2e299955f9d591d9013d1ebba36eeab07797e9a8
    --backend=true

services:
EOF

for ((i=0; i<NODE_COUNT; i++)); do
  cat >> docker-compose.yml <<EOF
  node${i}:
    <<: *node-defaults
    container_name: node${i}
    environment:
      NODE_ID: ${i}
      NODE_IP: $((START_IP + i))
EOF
done

cat >> docker-compose.yml <<EOF

networks:
  customnet:
    driver: bridge
    ipam:
      config:
        - subnet: 172.28.0.0/16
EOF

echo "生成完成！使用 docker-compose up -d 启动集群"