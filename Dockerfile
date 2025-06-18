# 第一阶段：构建Go二进制文件
# 使用带有所需Go版本的Alpine基础镜像以减小体积
FROM golang:1.23-alpine AS builder

# 安装 C 编译环境和依赖
RUN apk add --no-cache gcc musl-dev linux-headers git

# 设置工作目录以避免路径冲突
WORKDIR /ablockchain

# 首先复制依赖文件以利用Docker缓存层
COPY go.mod go.sum ./

# 下载依赖（国内可能需要在"go mode download"前增加"GOPROXY=https://goproxy.cn,direct GOSUMDB=off"）
RUN GOPROXY=https://goproxy.cn,direct GOSUMDB=off go mod download

# 复制项目所有源代码到容器
COPY . .

# 编译静态二进制文件
RUN go build -o /ablockchain/ablockchain ./main.go

# ==== 构建前端 ====
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend ./
RUN npm run build


# 第二阶段：构建轻量级生产镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \   
    tzdata          

# 设置时区（可选，按需调整）
ENV TZ=Asia/Shanghai

RUN apk add bash

RUN apk add curl nodejs npm

# 创建工作目录并确保权限正确
WORKDIR /ablockchain

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /ablockchain/ablockchain .
COPY --from=builder /ablockchain/genesis.json .
COPY --from=frontend-builder /frontend/dist /frontend


# 暴露应用程序使用的端口（根据实际情况修改）
EXPOSE 32000
EXPOSE 8545
EXPOSE 5137

COPY entrypoint.sh /ablockchain/entrypoint.sh
RUN chmod +x /ablockchain/entrypoint.sh

# 默认值，docker运行时环境变量可以覆盖是谁、、
ENV API_URL=http://localhost:8545

# 启动应用程序
ENTRYPOINT ["/ablockchain/entrypoint.sh"]
