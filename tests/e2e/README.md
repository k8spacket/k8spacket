### e2e tests for various Linux kernels

- e2e tests module uses `qemu` emulation tool to emulate various linux kernels, see target `start_qemu` in [Makefile](../../Makefile)
- `qemu` uses precompiled linux kernels, see [kernel compilation README](./vm/kernels/README.md)
- `qemu` uses prepared filesystem based on [Dockerfile](./vm/filesystem/Dockerfile)