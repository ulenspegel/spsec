.PHONY: build-linux clean

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o server main.go

clean:
	rm -f server
