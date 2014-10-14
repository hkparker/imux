Multiplexity
============

Multiplexity is an inverse multiplexer for file transfers.  Files are split into chunks that are inverse multiplexed over TCP sockets.  These sockets can exist on an arbitrary number of networks and each network and have an arbitrary number of sockets.  Using multiple sockets on a single network can improve performance and evade some implementations of traffic shaping / throttling while using multiple networks allows one to maximize bandwidth consumption on each network.  Multiplexity supports a number of other options as well, including adding more sockets mid transfer, resetting each connection after each chunk, and changing the chunk size.

Current status
--------------

Starting to implement in Go.

License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.

Contact
-------

Please feel free to contact me at haydenkparker@gmail.com
