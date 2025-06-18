
consensus

core
- blockchain
- account
- block
- transaction
- txpool

storage （s）

p2p （h）  libp2p

crypto （s）

networks

cli

utils



## tmux操作:

* **查看列表**
```bash
tmux ls
```
输出:
```
node1: 1 windows (created Mon Mar 25 10:00:00 2024)
node2: 1 windows (created Mon Mar 25 10:00:05 2024)
```

* **进入终端**
```bash
tmux attach-session -t node1
# 或简写:
tmux a -t node1
```

* **退出而不终止**
```
Ctrl + B，然后按 D
```

* **退出且终止**
```
Ctrl + C
```

* **关闭节点**
```
tmux kill-session -t node1
tmux kill-session -t node2

# 关闭所有会话
tmux kill-server
```

## Docker Usage

* **创建镜像**

在Dockerfile目录下，运行
```sh
# 创建名为 ablockchain:0.0.0 的镜像
# Dockerfile路径为当前路径 "."
docker build -t ablockchain:0.0.0 .
```

* **运行容器**

```sh
# 创建容器
docker run -d --name node0 -p 32000:32000 ablockchain:0.0.0

docker run -d --name node1 -p 32001:32000 ablockchain:0.0.0
```

*  **进入容器shell**

```sh
docker exec -it node0 /bin/bash
```

*  **停止和开始容器**

```sh
docker stop node0
docker start node0
```

*  **运行ablockchain**

进入容器后，在ablockchain路径下，启动可执行文件，并附加所需的参数

```sh
# 采用0.0.0.0时，监听地址会自动增加本地环回地址，以及其他地址，如网卡IP，docker IP等（推荐使用该种启动方式）
./ablockchain --listen /ip4/0.0.0.0/tcp/32000

# 这里listen的是docker容器的IP地址（不推荐，需要事先获知docker地址）
./ablockchain --listen /ip4/172.17.0.2/tcp/32000
./ablockchain --listen /ip4/172.17.0.3/tcp/32000
```

* **连接节点**

需要获知需连接节点的IP和ID
```sh
connect /ip4/172.17.0.3/tcp/32000/p2p/12D3KooWNR7FuHdJ6162Zbij55t4rxisBXfyYYHqebPqqkzdT4Js
```


*  **docker-compose**

``` sh
# 启动容器
docker-compose up -d

# 停止并删除docker-compose启动的容器，以及docker网络
docker-compose down

# 停止docker-compose启动的容器
docker-compose stop

# 删除docker-compose启动的容器
docker-compose rm

# 重新开启被停止而未删除的docker-compose启动的容器
docker-compose start
```

* **进入交互界面：**

```sh
docker exec -it node1 ./ablockchain attach
```

* **查看节点日志**

```sh
docker logs -f node1
```

## 自动化指定节点数量和生成Docker网络
运行shell文件生成Dockerfile
``` sh
bash generate_compose.sh
# 随后输入网络中的节点数量
```
运行Docker-compose启动网络
``` sh
docker-compose up -d
```

## 前端

1. 本地环境安装Node.js
2. 进入前端路径，即frontend，运行 npm install 下载安装所需的依赖
3. 在frondend路径下，运行如 npm run dev 开启前端服务器
