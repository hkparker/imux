Multiplexity
============

Multiplexity is an inverse multiplexer for file transfers.  Data is sent over an arbitrary number of TLS sockets routed over an arbitrary number of networks.  Using multiple sockets on a single network can improve performance and evade some implementations of traffic shaping / throttling while using multiple networks allows one to maximize bandwidth consumption on each network.  Multiplexity supports some other options as well, including adjusting the socket count mid transfer, resetting each connection after a n bytes, and directory transfers.

Current status
--------------

Starting to implement in Go.  The ReadQueue (File -> Chunks) and Buffer (Chunks -> File) objects are working well, I am now writing the IMUXSocket object, which passes Chunks over TLS sockets.

License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.

Contact
-------

Please feel free to contact me at hayden@hkparker.com
