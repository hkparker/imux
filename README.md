# imux

This is a go library and corresponding command line tool for inverse multiplexing sockets.

## Example

Let's say you wanted to expose an FTP server over imux.

On the server

```
imux -server --listen=0.0.0.0:443 --dial=127.0.0.1:21
```

Then on a client, inverse multiplex over 10 sockets

```
imux -client --binds={"0.0.0.0": 10} --listen=localhost:21 --dial=server:443
```

The client will use Trust Of First Use (TOFU) and TLS.  Now on the client, connect to localhost to FTP to the sever over the imux connection

```
ftp localhost
```
