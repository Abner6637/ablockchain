#!/bin/bash

# ========== Step 1: 注入 API 地址到前端 ==========
if [ -n "$API_URL" ]; then
  echo "[entrypoint] Injecting API_URL=$API_URL into frontend files..."
  sed -i "s|__API_URL_PLACEHOLDER__|${API_URL}|g" /frontend/index.html
  grep '__API_URL_PLACEHOLDER__' /frontend/index.html && echo "[entrypoint] Warning: API_URL placeholder not replaced!"
else
  echo "[entrypoint] No API_URL provided, skipping injection."
fi

# ========== Step 2: 启动前端服务器 ==========
echo "[entrypoint] Starting frontend at port 5173..."
npx http-server /frontend -p 5173 &

# ========== Step 3: 后端启动或阻塞 ==========
# 如果没有传入参数，就 tail 阻塞防止容器退出
if [ $# -eq 0 ]; then
  echo "[entrypoint] No command passed, holding container with tail -f /dev/null"
  tail -f /dev/null
else
  echo "[entrypoint] Executing command: $@"
  exec "$@"
fi