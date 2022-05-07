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
RUN g++ ./third_party/vfk_uoj_sandbox/run_program.cpp -o ./prebuilt/uoj_run -O2 -Wall
RUN g++ ./third_party/testlib/fcmp.cpp -o ./prebuilt/fcmp -O2 -Wall
RUN DEBIAN_FRONTEND=noninteractive apt-get remove -y git tzdata build-essential


# 运行时 镜像
FROM ubuntu:latest AS runner
ENV TZ=Asia/Shanghai

# 换阿里源
ARG CHANGE_SOURCE=true
RUN if [ ${CHANGE_SOURCE} = true ]; then \
    sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list && \
    sed -i 's/security.ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list \
;fi
RUN apt-get clean

RUN DEBIAN_FRONTEND=noninteractive apt-get update -y && apt-get upgrade -y
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && update-ca-certificates
# dns
RUN echo "hosts: files dns" > /etc/nsswitch.conf
# judger dependencies
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata zip unzip wget curl lsb-core bash bash-doc bash-completion build-essential gcc make g++ python2.7 python3 openjdk-8-jdk openjdk-11-jdk
RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone
# enter bash
RUN /bin/bash
# install java17
RUN mkdir -p /downloads
RUN wget -q -O /downloads/jdk-17_linux-x64_bin.tar.gz https://download.oracle.com/java/17/latest/jdk-17_linux-x64_bin.tar.gz
RUN tar -zxf /downloads/jdk-17_linux-x64_bin.tar.gz -C /usr/lib/jvm
# install java14
RUN wget -q -O /downloads/openjdk-14+36_linux-x64_bin.tar.gz https://download.java.net/openjdk/jdk14/ri/openjdk-14+36_linux-x64_bin.tar.gz
RUN tar -zxf /downloads/openjdk-14+36_linux-x64_bin.tar.gz -C /usr/lib/jvm/
RUN mv /usr/lib/jvm/jdk-17.0.3.1 /usr/lib/jvm/java-17-oracle-amd64
RUN mv /usr/lib/jvm/jdk-14 /usr/lib/jvm/java-14-openjdk-amd64
# check jvm dir
RUN ls -al /usr/lib/jvm
RUN ls /usr/lib/jvm/java-17-oracle-amd64
RUN ls /usr/lib/jvm/java-14-openjdk-amd64
# install golang
RUN wget -q -O /downloads/go1.18.1.linux-amd64.tar.gz https://golang.google.cn/dl/go1.18.1.linux-amd64.tar.gz
RUN tar -zxf /downloads/go1.18.1.linux-amd64.tar.gz -C /usr/local/
RUN /usr/local/go/bin/go version
# install nodejs
ENV NODE_KEYRING=/usr/share/keyrings/nodesource.gpg
ENV NODE_VERSION=node_18.x
ENV NODE_DISTRO=focal
RUN curl -fsSL https://deb.nodesource.com/gpgkey/nodesource.gpg.key | gpg --dearmor | tee "$NODE_KEYRING" >/dev/null
RUN gpg --no-default-keyring --keyring "$NODE_KEYRING" --list-keys
RUN echo "deb [signed-by=$NODE_KEYRING] https://deb.nodesource.com/$NODE_VERSION $NODE_DISTRO main" | tee /etc/apt/sources.list.d/nodesource.list
RUN echo "deb-src [signed-by=$NODE_KEYRING] https://deb.nodesource.com/$NODE_VERSION $NODE_DISTRO main" | tee -a /etc/apt/sources.list.d/nodesource.list
RUN apt-get update
RUN apt-get install nodejs 
RUN /usr/bin/node -v
# install rust and config
RUN curl https://sh.rustup.rs -sSf > /downloads/rustup-init.sh \
    && chmod +x /downloads/rustup-init.sh \
    && sh /downloads/rustup-init.sh -y
ENV PATH "$PATH:~/.cargo/bin"
RUN ~/.cargo/bin/rustc -V
RUN ~/.cargo/bin/cargo -V
RUN ~/.cargo/bin/rustup -V
# RUN bash -c "echo source $HOME/.cargo/env >> /etc/bash.bashrc"
# Update the local crate index
RUN ~/.cargo/bin/cargo search
# Install nightly rust.
RUN ~/.cargo/bin/rustup install nightly
# Initialize a dummy project so that we can pre-download the latest versions of
# the most popular crates, by building and destroy an empty project.
RUN mkdir /tmp/dummy-crate
WORKDIR /tmp/dummy-crate
RUN ~/.cargo/bin/cargo init \
    && echo "argparse = '*'" >> Cargo.toml \
    && echo "env_logger = '*'" >> Cargo.toml \
    && echo "hyper = '*'" >> Cargo.toml \
    && echo "itertools = '*'" >> Cargo.toml \
    && echo "log = '*'" >> Cargo.toml \
    && echo "matches = '*'" >> Cargo.toml \
    && echo "num = '*'" >> Cargo.toml \
    && echo "rand = '*'" >> Cargo.toml \
    && echo "regex = '*'" >> Cargo.toml \
    && echo "semver = '*'" >> Cargo.toml \
    && echo "tempdir = '*'" >> Cargo.toml \
    && echo "time = '*'" >> Cargo.toml \
    && echo "url = '*'" >> Cargo.toml \
    && ~/.cargo/bin/cargo build \
    && rm -rf /tmp/dummy-crate


RUN rm -r /downloads
RUN apt-get autoclean 
RUN apt-get autoremove 

# 创建文件夹
RUN mkdir -p /unilab-backend

WORKDIR /unilab-backend
COPY --from=builder /unilab-backend/main /unilab-backend/main
COPY --from=builder /unilab-backend/prebuilt /unilab-backend/prebuilt
COPY --from=builder /unilab-backend/conf.ini /unilab-backend/conf.ini
RUN chmod +x /unilab-backend/main

# 设置容器的启动命令，CMD是设置容器的启动指令
ENTRYPOINT [ "./main" ]
