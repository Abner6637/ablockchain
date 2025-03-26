
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
