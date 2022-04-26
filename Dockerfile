FROM golang:latest AS builder

ENV GO111MODULE='on'
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB='off'
ENV TZ=Asia/Shanghai

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y git tzdata build-essential
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone
# 创建文件夹
RUN mkdir /unilab-backend

# 设置工作目录
WORKDIR /unilab-backend

# 将主机路径.拷贝到镜像路径/app下
ADD . /unilab-backend

# 因为已经是在 /app下了，所以使用  ./
RUN mkdir -p ./prebuilt
RUN go build -v -o main .
RUN g++ ./third_party/vfk_uoj_sandbox/run_program.cpp -o ./prebuilt/uoj_run -O2
RUN g++ ./third_party/testlib/fcmp.cpp -o ./prebuilt/fcmp -O2
RUN DEBIAN_FRONTEND=noninteractive apt-get remove -y git tzdata build-essential


# 运行时 镜像
FROM ubuntu:latest AS runner
ENV TZ=Asia/Shanghai

RUN apt-get update && apt-get upgrade
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && update-ca-certificates
# dns
RUN echo "hosts: files dns" > /etc/nsswitch.conf
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata bash bash-doc bash-completion build-essential
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone
RUN apt-get autoclean 
RUN apt-get autoremove 
RUN /bin/bash
# 创建文件夹
RUN mkdir -p /unilab-backend

WORKDIR /unilab-backend
COPY --from=builder /unilab-backend/main /unilab-backend/main
COPY --from=builder /unilab-backend/prebuilt /unilab-backend/prebuilt
COPY --from=builder /unilab-backend/conf.ini /unilab-backend/conf.ini
RUN chmod +x /unilab-backend/main

# 暴露的端口
# EXPOSE 1323

# #设置容器的启动命令，CMD是设置容器的启动指令
ENTRYPOINT [ "./main" ]
