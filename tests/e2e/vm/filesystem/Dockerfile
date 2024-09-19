FROM ubuntu:22.04

RUN apt-get update \
    && apt-get install --no-install-recommends --no-install-suggests -y systemd \
    && apt-get install -y openssh-server \
    && apt-get install --no-install-recommends --no-install-suggests -y net-tools \
    && rm -rf /var/lib/apt/lists/*

RUN cd /lib/systemd/system && ln -sf multi-user.target default.target

RUN systemctl enable getty@ttyS0.service
RUN systemctl enable ssh.service

RUN echo "root:root" | chpasswd

RUN sed -i 's/ExecStart=.*/ExecStart=-\/sbin\/agetty --noissue --autologin root %I $TERM/g' /lib/systemd/system/getty@.service
RUN sed -i 's/TTYVTDisallocate=yes/TTYVTDisallocate=no/g' /lib/systemd/system/getty@.service
RUN sed -i 's/.*PermitRootLogin.*/PermitRootLogin yes/g' /etc/ssh/sshd_config

RUN echo "k8spacket" > /etc/hostname

COPY key.pem /root/
COPY cert.pem /root/
COPY https_server.py /root/

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
    ExecStart=/bin/bash -c 'python3.10 /root/https_server.py'

    [Install]
    WantedBy=multi-user.target
EOF

RUN systemctl enable eth0.service

RUN cat <<EOF >> /etc/systemd/system/k8spacket.conf
    K8S_PACKET_TCP_LISTENER_PORT=6676
    K8S_PACKET_TLS_CERTIFICATE_CACHE_TTL=30s
    K8S_PACKET_TCP_LISTENER_INTERFACES_COMMAND="echo -n eth0,lo"
    K8S_PACKET_TCP_LISTENER_INTERFACES_REFRESH_PERIOD=3s
    K8S_PACKET_K8S_RESOURCES_DISABLED=true
EOF

RUN cat <<EOF >> /etc/systemd/system/k8spacket.service
    [Unit]
    Description=k8spacket service

    [Service]
    EnvironmentFile=/etc/systemd/system/k8spacket.conf
    User=root
    WorkingDirectory=/root
    Type=idle
    ExecStart=/root/k8spacket

    [Install]
    WantedBy=multi-user.target
EOF