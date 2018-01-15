config:
	mkdir rootfs
	tar -xvf -C rootfs

build: 
	GOOS=linux GOARCH=amd64 go build

run: 
	sudo ./container run /bin/bash 