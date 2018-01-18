# container
simple container implementation in go.

## network 
set dns resolver in container:
```sh
echo "nameserver 8.8.8.8" >> /etc/resolv.conf
```
set ip forward in host:
```sh 
sysctl -w net.ipv4.ip_forward=1
```
