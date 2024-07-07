FROM ubuntu:22.04 AS libbpf

RUN apt-get update && apt-get install -y libelf-dev libpcap-dev libbfd-dev binutils-dev build-essential make
RUN apt-get install -y linux-tools-common git curl

RUN mkdir /home/k8spacket/
WORKDIR /home/k8spacket

# install bpftool from the source cases by error `bpftool not found for kernel v...`
# https://github.com/lizrice/lb-from-scratch/issues/1#issuecomment-1537098872
RUN rm -f /usr/sbin/bpftool \
    && git clone --recurse-submodules https://github.com/libbpf/bpftool.git \
    && cd ./bpftool/src \
    && make install \
    && cd ../..

COPY ./libbpf.sh .
RUN ./libbpf.sh


FROM golang:1.22.5 AS build

RUN export DEBIAN_FRONTEND=noninteractive && apt-get update && apt-get install -y clang llvm

RUN mkdir /home/k8spacket/

COPY ./broker /home/k8spacket/broker
COPY ./ebpf /home/k8spacket/ebpf
COPY ./log /home/k8spacket/log
COPY ./plugins /home/k8spacket/plugins
COPY ./go.mod /home/k8spacket/
COPY ./go.sum /home/k8spacket/
COPY ./init.sh /home/k8spacket/
COPY *.go /home/k8spacket/

COPY --from=libbpf ./home/k8spacket/*.h /home/k8spacket/ebpf/inet/bpf
#`-ldflags "-w -s"` means strip the debugging information to make binary smaller
RUN cd /home/k8spacket/ebpf/inet && go generate -ldflags "-w -s"

COPY --from=libbpf ./home/k8spacket/*.h /home/k8spacket/ebpf/tc/bpf
RUN cd /home/k8spacket/ebpf/tc && go generate -ldflags "-w -s"

RUN cd /home/k8spacket/ && ./init.sh


FROM alpine:3.20.1

RUN apk add --no-cache iproute2 libc6-compat

RUN mkdir /home/k8spacket && cd /home/k8spacket
WORKDIR /home/k8spacket

COPY --from=build ./home/k8spacket/k8spacket /home/k8spacket/
COPY ./fields.json /home/k8spacket/
COPY ./GeoLite2-City.mmdb /home/k8spacket/

# need to run as root regarding the use of kernel tracing info `/sys/kernel/tracing`
CMD ["./k8spacket"]
