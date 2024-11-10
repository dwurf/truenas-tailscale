all: test build

test:
	go test ./...

build:
	GOOS=linux GOARCH=amd64 go build -o dist/truenas-tailscale .

clean:
	rm dist/truenas-tailscale
	rmdir dist
