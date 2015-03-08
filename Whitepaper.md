Multiplexity
============

Motivation
----------

Originally this project started because I wanted to download a file over multiple internet connections.  I imagined a server that would accent n sockets and inverse multiplex transfers between these sockets.  If a client had multiple internet connections, source based routing could be used to bind some sockets to certain interfaces.  I began writing a simple version of this idea in Ruby, and tested it in a VM with multiple virtual interfaces.

Design
------


Hosts and Sessions
------------------

A host is a computer you wish to transfer files to or from.  This means your local host

TransferGroups
--------------

TransferGroups manage moving a file over a collection of sockets.

TransferSockets
---------------

A TransferSocket is a class that owns a TLS socket and 


Chunks, ReadQueues, and WriteBuffers
------------------------------------

