Multiplexity
============

Multiplexity is an inverse multiplexer for file transfers.  Data is sent over an arbitrary number of TLS sockets routed over an arbitrary number of networks.  Using multiple sockets on a single network can improve performance and evade some implementations of traffic shaping / throttling while using multiple networks allows one to combine the bandwidth consumption of each network.  Multiplexity can also reset the transfer sockets after n bytes of data to appear as a large number of small TLS sessions.

Installation
------------

### Dependencies ###

Server

	* Linux (requires unix sockets and authenticates against /etc/shadow)
	* github.com/kless/osutil/user/crypt/sha512_crypt
	* github.com/twinj/uuid

Client

	* gtk

### Building the server ###

go build Session.go
go build Server.go

Running the server is as simple as:
./Server

### Building the client ###

go build Multiplexity.go
./Multiplexity

Usage
-----

Connect to a host, create a queue

License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.

Contact
-------

Feel free to contact me at hayden@hkparker.com
