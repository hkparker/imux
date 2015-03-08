Multiplexity
============

Multiplexity is an inverse multiplexer for file transfers.  Data is sent over an arbitrary number of TLS sockets routed over an arbitrary number of networks.  Using multiple sockets on a single network can improve performance and evade some implementations of traffic shaping / throttling while using multiple networks allows one to maximize bandwidth consumption on each network.  Multiplexity supports some other options as well, including resetting each connection after every chunk of data, changing the chunk size, and increasing the socket count during a session.

Current status
--------------

I've experiemented with several ideas and protocol features in a Ruby proof of concept implementation, and am now in the process of formalizing everything in Go.  I expect to have the basic API complete in the next few months.

License
-------

This project is licensed under the MIT license, please see LICENSE.md for more information.

Contact
-------

Feel free to contact me at hayden@hkparker.com
