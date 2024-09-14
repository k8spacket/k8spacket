FROM ubuntu:22.04

RUN apt-get update
RUN apt-get install -y systemd
RUN apt-get install -y openssh-server
RUN apt-get install -y ifupdown net-tools

RUN apt-get install -y nginx
RUN apt-get install -y curl
RUN apt-get install -y vim

RUN cd /lib/systemd/system && ln -sf multi-user.target default.target

RUN systemctl enable getty@ttyS0.service
RUN systemctl enable ssh.service

RUN echo "root:root" | chpasswd

RUN sed -i 's/ExecStart=.*/ExecStart=-\/sbin\/agetty --noissue --autologin root %I $TERM/g' /lib/systemd/system/getty@.service
RUN sed -i 's/TTYVTDisallocate=yes/TTYVTDisallocate=no/g' /lib/systemd/system/getty@.service
RUN sed -i 's/.*PermitRootLogin.*/PermitRootLogin yes/g' /etc/ssh/sshd_config


COPY k8spacket /root/
COPY entrypoint.sh /root/
RUN chmod +x /root/entrypoint.sh

RUN echo "k8spacket" > /etc/hostname

COPY k8spacket.conf /etc/nginx/conf.d
COPY key.pem /etc/nginx/conf.d
COPY cert.pem /etc/nginx/conf.d

RUN cat <<EOF >> /etc/systemd/system/eth0.service
    [Unit]
    Description=eth0 service

    [Service]
    User=root
    WorkingDirectory=/root
    Type=oneshot
    ExecStart=ifconfig eth0 10.0.2.15 netmask 255.255.255.0
    ExecStart=route add default gw 10.0.2.2
    ExecStart=/bin/bash -c '/usr/bin/echo nameserver 8.8.8.8 > /etc/resolv.conf'
    ExecStart=/bin/bash -c '/usr/bin/echo "127.0.0.1 k8spacket.domain" > /etc/hosts'

    [Install]
    WantedBy=multi-user.target
EOF

RUN systemctl enable eth0.service

RUN cat <<EOF >> /etc/systemd/system/k8spacket.service
    [Unit]
    Description=k8spacket service

    [Service]
    User=root
    WorkingDirectory=/root
    Type=idle
    ExecStart=/bin/bash /root/entrypoint.sh

    [Install]
    WantedBy=multi-user.target
EOF

RUN systemctl enable k8spacket.service