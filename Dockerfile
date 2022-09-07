FROM --platform=linux/amd64 golang:1.19.1 AS build

RUN export DEBIAN_FRONTEND=noninteractive && apt-get update && apt-get install -y libpcap-dev

RUN mkdir /home/k8spacket/

COPY ./k8s /home/k8spacket/k8s
COPY ./metrics /home/k8spacket/metrics
COPY ./tcp /home/k8spacket/tcp
COPY ./tools /home/k8spacket/tools
COPY *.go /home/k8spacket/
COPY ./init.sh /home/k8spacket/
COPY ./go.mod /home/k8spacket/
COPY ./go.sum /home/k8spacket/

RUN cd /home/k8spacket/ && ./init.sh

FROM golang:alpine

RUN apk update && apk add libpcap-dev libcap net-tools iproute2 libc6-compat

RUN cd /usr/lib/ && ln -s libpcap.so libpcap.so.0.8 && cd -

RUN addgroup -S k8spacket
RUN adduser --disabled-password --gecos "" --home /home/k8spacket --ingroup k8spacket k8spacket

WORKDIR /home/k8spacket

COPY --from=build ./home/k8spacket/k8spacket /home/k8spacket/
COPY ./fields.json /home/k8spacket/

RUN chown k8spacket:k8spacket /home/k8spacket/*

RUN chgrp k8spacket /home/k8spacket/k8spacket && chmod 750 /home/k8spacket/k8spacket

RUN setcap cap_net_raw,cap_net_admin=ep /home/k8spacket/k8spacket

USER k8spacket

CMD ["./k8spacket"]
