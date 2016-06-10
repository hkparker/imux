imux
====

imux is an inverse multiplexer for file transfers.  Data is sent over an arbitrary number of TLS sockets routed over an arbitrary number of networks.  Using multiple sockets on a single network can improve TCP performance and evade some implementations of traffic shaping / throttling while using multiple networks allows one to combine the bandwidth each network provides.  The client authenticates server certificates using trust of first use (TOFU), similar to SSH.

Installation
------------

```
go get github.com/hkparker/imux
```

Usage
-----

Start a server
```
imux
```

Connect
```
imux --host=example.com
```

Help
```
Usage of imux:
  -bind string
	address to bind an imux server on (default "0.0.0.0:443")
  -cert string
	pem file with certificate to present when in server mode, auto generated (default "cert.pem")
  -chunk int
	size of each file chunk in bytes, specified by the client (default 5242880)
  -daemon
	run the server in the background
  -host string
	imux server to connect to
  -key string
	pem file with key for certificate presented in server mode, auto generated (default "key.pem")
  -networks string
	socket configuration string for clients: <bind ip>:<count>; (default "0.0.0.0:200")
  -recycle int
	bytes transferred before client closes and replaces socket, default unlimited
  -session string
	unix socket used for IPC, set by server for user context processes
  -user string
	username (default "example")
```

License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.
