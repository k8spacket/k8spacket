DOCKER_BUILDKIT=1 docker build --output "type=tar,dest=filesystem.tar" .

virt-make-fs --format=qcow2 --size=+100M filesystem.tar filesystem-large.qcow2
qemu-img convert filesystem-large.qcow2 -O qcow2 filesystem.qcow2
rm filesystem-large.qcow2