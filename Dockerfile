# 第一阶段：构建Go二进制文件
# 使用带有所需Go版本的Alpine基础镜像以减小体积
FROM golang:1.23-alpine AS builder

# 设置工作目录以避免路径冲突
WORKDIR /app

# 首先复制依赖文件以利用Docker缓存层
COPY go.mod go.sum ./

# 下载依赖（若go.mod未更改则跳过此步骤）
RUN GOPROXY=https://goproxy.cn,direct GOSUMDB=off go mod download

# 复制项目所有源代码到容器
COPY . .

# 编译静态二进制文件（禁用CGO，适配Alpine）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/myapp ./main.go

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
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/myapp .
COPY --from=builder /app/genesis.json .


# 暴露应用程序使用的端口（根据实际情况修改）
EXPOSE 32000

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# 启动应用程序
ENTRYPOINT ["/app/entrypoint.sh"]
