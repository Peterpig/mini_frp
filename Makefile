all: frps frpc

frps:
	go build -o bin/frps ./cmd/frps

frpc:
	go build -o bin/frpc ./cmd/frpc
