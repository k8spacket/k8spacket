FROM golang:1.21.5 AS build

RUN export DEBIAN_FRONTEND=noninteractive && apt-get update && apt-get install -y libpcap-dev

RUN mkdir /home/k8spacket/

COPY ./broker /home/k8spacket/broker
COPY ./plugins /home/k8spacket/plugins
COPY ./tcp /home/k8spacket/tcp
COPY ./tools /home/k8spacket/tools
COPY *.go /home/k8spacket/
COPY ./init.sh /home/k8spacket/
COPY ./go.mod /home/k8spacket/
COPY ./go.sum /home/k8spacket/

RUN cd /home/k8spacket/ && ./init.sh

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y libcap2-bin libpcap0.8 iproute2

RUN useradd --create-home k8spacket

WORKDIR /home/k8spacket

COPY --from=build ./home/k8spacket/k8spacket /home/k8spacket/
COPY ./fields.json /home/k8spacket/

RUN chown k8spacket:k8spacket /home/k8spacket/*

RUN chgrp k8spacket /home/k8spacket/k8spacket && chmod 750 /home/k8spacket/k8spacket

RUN setcap cap_net_raw,cap_net_admin=ep /home/k8spacket/k8spacket

USER k8spacket

CMD ["./k8spacket"]
