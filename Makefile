config:
	mkdir rootfs
	tar -xvf -C rootfs

build: 
	go build main.go 

run: sudo ./container run /bin/bash 