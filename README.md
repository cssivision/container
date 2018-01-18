# container
simple container implementation in go.

## network 
set dns resolver in `/etc/resov.conf`:
```sh
echo "nameserver 8.8.8.8" >> /etc/resov.conf
```
set ip forward:
```sh 
sysctl -w net.ipv4.ip_forward=1
```
