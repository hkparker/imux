# imux

imux is a go library and corresponding command line tool for inverse multiplexing sockets

![imux](https://cloud.hkparker.com/index.php/apps/files_sharing/ajax/publicpreview.php?x=2560&y=971&a=true&file=imux.png&t=QpoYNKvRQUL96ig&scalingup=0)

An imux client will create a listener and forward data from any connections to that listener to an imux server, using a configurable number of sockets.  An imux server receives data and opens corresponding sockets to the final destination.  Encryption is done with self-signed TLS and Trust Of First Use (TOFU).

## example

let's say you wanted to expose an SSH server over imux

**server**

serve imux on `0.0.0.0:443` and connect out to `localhost:22`

```
imux -server --listen=0.0.0.0:443 --dial=localhost:22
```

**client**

inverse multiplex over 10 sockets bound to any interface and connect to the server

```
imux -client --binds='{"0.0.0.0": 10}' --listen=localhost:22 --dial=server:443
```

now on the client, connect to `localhost:22` to ssh to the sever's `localhost:22` over the imux connection

```
ssh localhost
```

## multiple routes

imux can be used to transport a single socket over multiple internet connections using source routing in linux

For example, consider simultaneously using two interfaces:

|Interface|Address|Default Gateway|
|:-------:|:-----:|:-------------:|
|`eth0`|`192.168.1.2`|`192.168.1.1`|
|`eth1`|`10.0.0.2`|`10.0.0.1`|

**create routing tables**

```
echo '128    imux0' >> /etc/iproute2/rt_tables
echo '129    imux1' >> /etc/iproute2/rt_tables
```

**add routes**

```
ip route add default via 192.168.1.1 table imux0 dev eth0
ip route add default via 10.0.0.1 table imux1 dev eth1
```

**add rules**

```
ip rule add from 192.168.1.2 table imux0
ip rule add from 10.0.0.2 table imux1
```

**flush cache**

```
ip route flush cache
```

**connect with binds**

here we choose 20 sockets on each interface

```
imux -client --binds='{"192.168.1.2": 20, "10.0.0.2": 20}' --listen=localhost:22 --dial=server:443
```
