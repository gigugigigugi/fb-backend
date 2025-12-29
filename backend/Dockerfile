# ---- Build Stage ----
# 使用官方的 Go 镜像作为构建器。我们选择 alpine 版本以减小构建环境的大小。
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 Go 模块文件。先复制这两个文件可以利用 Docker 的缓存机制，
# 只要它们不改变，下一步的依赖下载就不会重复执行。
COPY go.mod go.sum ./

# 下载所有依赖项
RUN go mod download

# 复制整个 backend 项目的源代码到容器中
COPY . .

# 编译 Go 应用。
# CGO_ENABLED=0 创建一个静态链接的二进制文件，不依赖任何 C 库。
# GOOS=linux 确保它能在 Cloud Run 的 Linux 环境下运行。
# -o /server 指定输出文件名为 'server'。
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /server ./cmd/main.go


# ---- Runtime Stage ----
# 使用一个极简的 "distroless" 镜像。它不包含 shell 或任何其他不必要的程序，
# 只包含运行我们程序所必需的内容，非常安全。
FROM gcr.io/distroless/static-debian11

# 从构建阶段复制编译好的二进制文件到当前镜像
COPY --from=builder /server /server

# Cloud Run 会通过这个环境变量告诉我们的应用应该监听哪个端口。
# 我们的 main.go 已经配置为读取这个变量。
ENV PORT 8080

# 容器启动时要执行的命令。
CMD ["/server"]
