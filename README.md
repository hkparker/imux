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
go build session.go types.go && go build imux-server.go types.go
./imux-server -help

go build imux.go types.go common.go read_queue.go
./imux -help
```

License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.
