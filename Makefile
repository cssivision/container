config:
	mkdir -p rootfs
	sudo tar -xvf rootfs.tar -C rootfs

build: 
	GOOS=linux GOARCH=amd64 go build

run: 
	sudo ./container run /bin/bash 