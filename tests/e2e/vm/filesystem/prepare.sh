# build filesystem image and store as tar archive
DOCKER_BUILDKIT=1 docker build --output "type=tar,dest=filesystem.tar" .

# convert tar to qcow2 image
virt-make-fs --format=qcow2 --size=+100M filesystem.tar filesystem-large.qcow2
# reduce size of image
qemu-img convert filesystem-large.qcow2 -O qcow2 filesystem.qcow2
rm filesystem-large.qcow2