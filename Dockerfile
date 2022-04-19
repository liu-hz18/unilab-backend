FROM golang:alpine AS builder

ENV GO111MODULE='on'
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB='off'
ENV TZ=Asia/Shanghai

RUN apk update
RUN apk --no-cache --virtual build-dependencies add git tzdata
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
# 创建文件夹
RUN mkdir /unilab-backend

# 设置工作目录
WORKDIR /unilab-backend

# 将主机路径.拷贝到镜像路径/app下
ADD . /unilab-backend

# 因为已经是在 /app下了，所以使用  ./
RUN go build -v -o main .
RUN apk del build-dependencies


# 运行时 镜像
FROM alpine:latest AS runner
ENV TZ=Asia/Shanghai
# 配置国内源
RUN echo "http://mirrors.aliyun.com/alpine/latest-stable/main/" > /etc/apk/repositories
RUN apk update
RUN apk --no-cache add ca-certificates && update-ca-certificates
# dns
RUN echo "hosts: files dns" > /etc/nsswitch.conf
RUN apk --no-cache add tzdata
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
# 创建文件夹
RUN mkdir -p /unilab-backend

WORKDIR /unilab-backend
COPY --from=builder /unilab-backend/main /unilab-backend/main
COPY --from=builder /unilab-backend/conf.ini /unilab-backend/conf.ini
RUN chmod +x /unilab-backend/main

# 暴露的端口
# EXPOSE 1323

# #设置容器的启动命令，CMD是设置容器的启动指令
ENTRYPOINT [ "./main" ]
