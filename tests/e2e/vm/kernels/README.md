### Kernel compilation steps

- download kernel from repo https://cdn.kernel.org/pub/linux/kernel/
- prepare configuration file
```shell
make defconfig # new config based on defaults for the current ARCH

./scripts/config --set-val CONFIG_DEBUG_INFO_DWARF_TOOLCHAIN_DEFAULT y # Rely on the default DWARF (debugging information file format) toolchain's version
./scripts/config --set-val CONFIG_BPF_SYSCALL y # Enable the bpf() system call that allows to manipulate BPF programs and maps in kernel
./scripts/config --set-val CONFIG_DEBUG_INFO y - # Compile kernel with debug info, autoset by CONFIG_DEBUG_INFO_DWARF_TOOLCHAIN_DEFAULT on 6.x kernels
./scripts/config --set-val CONFIG_DEBUG_INFO_BTF y # Generate BTF type information from DWARF debug info

#./scripts/config --set-val CONFIG_NET_CLS_ACT y # enable network classification, autoset by CONFIG_NET_CLS_BPF
./scripts/config --set-val CONFIG_NET_CLS_BPF y # BPF-based classifier
./scripts/config --set-val CONFIG_NET_SCH_INGRESS y # Ingress/classifier-action qdisc

make olddefconfig # regenerate config based on new values and defaults
```
- build linux kernel
```shell
make -j8
```