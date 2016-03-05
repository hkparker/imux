imux
====

imux is an inverse multiplexer for file transfers.  Data is sent over an arbitrary number of TLS sockets routed over an arbitrary number of networks.  Using multiple sockets on a single network can improve TCP performance and evade some implementations of traffic shaping / throttling while using multiple networks allows one to combine the bandwidth each network provides.  The client authenticates server certificates using trust of first use (TOFU), similar to SSH.

Building
--------

Deps:

```
make deps
```

Build:

```
make
```

Usage
-----

```
Usage of ./imux:
  -chunksize int
    	size of each file chink in byte (default 5242880)
  -host string
    	hostname
  -networks string
    	socket configuration string: <bind ip>:<count>; (default "0.0.0.0:200")
  -port int
    	port (default 443)
  -user string
    	username (default "demo")

Usage of ./imux-server:
  -cert string
    	pem file with certificate to present (default "ca.pem")
  -key string
    	pem file with key for certificate (default "ca.key")
  -listen string
    	address to listen on (default "0.0.0.0")
  -port int
    	port to listen on (default 443)
```

License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.
