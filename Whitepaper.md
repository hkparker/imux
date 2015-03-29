Multiplexity
============

Motivation
----------

Originally this project started because I wanted to download a file over multiple internet connections.  I imagined a server that would accept n sockets and inverse multiplex transfers over these sockets.  If a client had multiple internet connections, source based routing could be used to bind some sockets to certain interfaces, and it wouldn't matter to the server which networks those sockets came from.  I began writing a simple version of this idea in Ruby, and tested it in a VM with multiple virtual interfaces.

Design
------


Hosts and Sessions
------------------



TransferGroups
--------------



TransferSockets
---------------



Chunks, ReadQueues, and WriteBuffers
------------------------------------

