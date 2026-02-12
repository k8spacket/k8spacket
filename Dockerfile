FROM ubuntu:24.04 AS libbpf

RUN apt-get update && apt-get install -y libelf-dev libpcap-dev libbfd-dev libssl-dev binutils-dev build-essential make
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


FROM golang:1.26.0 AS build

RUN export DEBIAN_FRONTEND=noninteractive && apt-get update && apt-get install -y clang llvm

RUN mkdir /home/k8spacket/

COPY ./cmd /home/k8spacket/cmd
COPY ./internal /home/k8spacket/internal
COPY ./go.mod /home/k8spacket/
COPY ./go.sum /home/k8spacket/

#`-ldflags "-w -s"` means strip the debugging information to make binary smaller
COPY --from=libbpf ./home/k8spacket/*.h /home/k8spacket/internal/ebpf/inet/bpf
RUN cd /home/k8spacket/internal/ebpf/inet && go generate -ldflags "-w -s"

COPY --from=libbpf ./home/k8spacket/*.h /home/k8spacket/internal/ebpf/tc/bpf
RUN cd /home/k8spacket/internal/ebpf/tc && go generate -ldflags "-w -s"
COPY --from=libbpf ./home/k8spacket/*.h /home/k8spacket/internal/ebpf/socketfilter/bpf
RUN cd /home/k8spacket/internal/ebpf/socketfilter && go generate -ldflags "-w -s"

RUN cd /home/k8spacket && env CGO_ENABLED=0 go build ./cmd/k8spacket


FROM gcr.io/distroless/static-debian12

WORKDIR /home/k8spacket

COPY --from=build ./home/k8spacket/k8spacket /home/k8spacket/
COPY ./fields.json /home/k8spacket/
#COPY ./GeoLite2-City.mmdb /home/k8spacket/

# need to run as root regarding the use of kernel tracing info `/sys/kernel/tracing`
CMD ["./k8spacket"]