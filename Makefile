session:
	go build imux-session/main.go

server: session
	go build imux-server/main.go

client:
	go build imux/main.go

all: server client

deps:
	go get github.com/hkparker/TLJ
	go get github.com/kless/osutil/user/crypt/sha512_crypt
	go get golang.org/x/crypto/ssh/terminal
	go get github.com/dustin/go-humanize
