# Container
simple container implementation in go.

## Material

1, Unprivileged containers on Go.
* [Part1: User and PID namespaces](http://lk4d4.darth.io/posts/unpriv1/)
* [Part2: UTS namespace (setup namespaces)](http://lk4d4.darth.io/posts/unpriv2/)
* [Part3: Mount namespace](http://lk4d4.darth.io/posts/unpriv3/)
* [Part4: Network namespace](http://lk4d4.darth.io/posts/unpriv4/)

2, Docker implemented in around 100 lines of bash.
* [https://github.com/p8952/bocker](https://github.com/p8952/bocker)

3, Code to accompany the "Namespaces in Go" series of articles.

* [Part 1: Linux Namespaces](https://medium.com/@teddyking/linux-namespaces-850489d3ccf)
* [Part 2: Namespaces in Go - Basics](https://medium.com/@teddyking/namespaces-in-go-basics-e3f0fc1ff69a)
* [Part 3: Namespaces in Go - User](https://medium.com/@teddyking/namespaces-in-go-user-a54ef9476f2a)
* [Part 4: Namespaces in Go - reexec](https://medium.com/@teddyking/namespaces-in-go-reexec-3d1295b91af8)
* [Part 5: Namespaces in Go - Mount](https://medium.com/@teddyking/namespaces-in-go-mount-e4c04fe9fb29)
* [Part 6: Namespaces in Go - Network](https://medium.com/@teddyking/namespaces-in-go-network-fdcf63e76100)
* [Part 7: Namespaces in Go - UTS](https://medium.com/@teddyking/namespaces-in-go-uts-d47aebcdf00e)

4, Shell script to create network namespace.e7c498b222f0aa49efa04b858e5432cda437fc19
* [https://github.com/cssivision/container/blob/master/network-namespace.sh](https://github.com/cssivision/container/blob/master/network-namespace.sh)

5, Iptables.
* [a-deep-dive-into-iptables-and-netfilter-architecture](https://www.digitalocean.com/community/tutorials/a-deep-dive-into-iptables-and-netfilter-architecture)
* [Security Guide IPTables](https://access.redhat.com/documentation/en-US/Red_Hat_Enterprise_Linux/6/html-single/Security_Guide/index.html#sect-Security_Guide-IPTables)


## Network.
set dns resolver in container:
```sh
echo "nameserver 8.8.8.8" >> /etc/resolv.conf
```
set ip forward in host:
```sh 
sysctl -w net.ipv4.ip_forward=1
```

## Run.
```sh
make config
make build
make run
```
