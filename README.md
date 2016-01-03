Multiplexity
============

Multiplexity is an inverse multiplexer for file transfers.  Data is sent over an arbitrary number of TLS sockets routed over an arbitrary number of networks.  Using multiple sockets on a single network can improve TCP performance and evade some implementations of traffic shaping / throttling while using multiple networks allows one to combine the bandwidth each network provides.  The client authenticates server certificates using trust of first use (TOFU), similar to SSH.

Installation
------------

### Dependencies ###

**Server**

* Linux (requires unix sockets and authenticates against /etc/shadow)
* `go get github.com/kless/osutil/user/crypt/sha512_crypt`
* `go get github.com/hkparker/TLJ`

**Client**

* `go get github.com/hkparker/TLJ`
* `go get golang.org/x/crypto/ssh/terminal`

### Building the server ###

`go build imux-server.go types.go`

**Usage:**

```
```

### Building the client ###

``

**Usage:**

```
```

Example
-------



License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.
