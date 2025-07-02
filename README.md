# 平台模块

## core

区块链的核心模块。

包含区块、交易、区块链本身、账户、交易池、EVM等。

## consensus

共识模块。

可运行的共识算法有PoW和BFT算法，其中BFT算法实现了PBFT和HotStuff。

## P2P

网络层的通信模块，主要负责各节点之间的P2P通信，如广播区块、共识报文等。

## frontend

前端界面模块。

## system

整个项目的总控制模块。负责控制整个系统的启动、命令行的交互等。

## crypto

加密模块。包含哈希计算、密钥生成、签名与签名验证、默克尔树等。

## Storage

存储模块。存储区块、账户状态等数据。

## event

事件模块。负责事件的处理。

## Config

创世节点（genesis block）的配置。

## Cli

命令行启动、命令行启动参数设置等。

## API

api方法与调用。

# 启动方式

## tmux启动

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

## Docker 启动

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

### 启动Docker网络

运行shell文件生成Dockerfile

``` sh
bash generate_compose.sh
# 随后输入网络中的节点数量
```

运行Docker-compose启动网络

``` sh
docker-compose up -d
```

## 前端的启动

1. 本地环境安装Node.js
2. 进入前端路径，即frontend，运行 npm install 下载安装所需的依赖
3. 在frondend路径下，运行如 npm run dev 开启前端服务器
4. 在浏览器中输入服务器的网址进入前端界面



