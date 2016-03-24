all: server client

session:
	go build -o imux-session imux-session.d/main.go

server: session
	go build -o imux-server imux-server.d/main.go

client:
	go build -o imux imux.d/main.go

deps:
	go get github.com/hkparker/TLJ
	go get github.com/kless/osutil/user/crypt/sha512_crypt
	go get golang.org/x/crypto/ssh/terminal
	go get github.com/dustin/go-humanize
	go get github.com/jbenet/go-base58
