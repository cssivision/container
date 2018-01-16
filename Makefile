config:
	mkdir rootfs
	sudo tar rootfs.tar -xvf -C rootfs

build: 
	go get -u -v  github.com/vishvananda/netlink
	GOOS=linux GOARCH=amd64 go build

run: 
	sudo ./container run /bin/bash 