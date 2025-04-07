#!/bin/bash

# cp /genesis.json /app/genesis.json

# 最后运行程序，带上任何命令行参数
# exec ./myapp 

# 如果没有传入参数，就 tail 阻塞防止容器退出
if [ $# -eq 0 ]; then
  echo "[entrypoint] No command passed, holding container with tail -f /dev/null"
  tail -f /dev/null
else
  echo "[entrypoint] Executing command: $@"
  exec "$@"
fi