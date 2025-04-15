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

# 第二阶段：构建轻量级生产镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \   
    tzdata          

# 设置时区（可选，按需调整）
ENV TZ=Asia/Shanghai

RUN apk add bash

# 创建工作目录并确保权限正确
WORKDIR /ablockchain

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /ablockchain/ablockchain .
COPY --from=builder /ablockchain/genesis.json .


# 暴露应用程序使用的端口（根据实际情况修改）
EXPOSE 32000

COPY entrypoint.sh /ablockchain/entrypoint.sh
RUN chmod +x /ablockchain/entrypoint.sh

# 启动应用程序
ENTRYPOINT ["/ablockchain/entrypoint.sh"]
