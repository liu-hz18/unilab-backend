FROM golang:latest AS builder

ENV GO111MODULE='on'
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB='off'

# 创建文件夹
RUN mkdir /unilab-backend

# 设置工作目录
WORKDIR /unilab-backend

# 将主机路径.拷贝到镜像路径/app下
ADD . /unilab-backend

# 因为已经是在 /app下了，所以使用  ./
RUN go build -v -o main .

EXPOSE 1323
ENTRYPOINT [ "./main" ]

# # 运行时 镜像
# FROM alpine:latest
# # 配置国内源
# RUN echo "http://mirrors.aliyun.com/alpine/latest-stable/main/" > /etc/apk/repositories
# RUN apk update
# RUN apk --no-cache add ca-certificates && update-ca-certificates
# # dns
# RUN echo "hosts: files dns" > /etc/nsswitch.conf
# # 创建文件夹（根据个人选择）
# RUN mkdir -p /unilab-backend

# WORKDIR /unilab-backend
# COPY --from=builder /unilab-backend/main /usr/bin/main
# RUN chmod +x /usr/bin/main

# # 暴露的端口
# EXPOSE 1323

# #设置容器的启动命令，CMD是设置容器的启动指令
# ENTRYPOINT ["main"]
