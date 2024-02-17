#!/usr/bin/env bash

set -x

# Version of libbpf to fetch headers from
LIBBPF_VERSION=1.3.0

# The headers we want
prefix=libbpf-"$LIBBPF_VERSION"
headers=(
    "$prefix"/src/bpf_core_read.h
    "$prefix"/src/bpf_endian.h
    "$prefix"/src/bpf_helper_defs.h
    "$prefix"/src/bpf_helpers.h
    "$prefix"/src/bpf_tracing.h
)

# Fetch libbpf release and extract the desired headers
curl -sL "https://github.com/libbpf/libbpf/archive/refs/tags/v${LIBBPF_VERSION}.tar.gz" | \
    tar -xz -C . --xform='s#.*/##' "${headers[@]}"

bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h