all: test build

test: test-unit test-integration

test-unit:
	go test ./...

test-integration:
	go install github.com/juanfont/headscale/cmd/headscale@v0.26.1
	(headscale serve -c .headscale/config.yml 2>/dev/null &)
	go test ./... -- -headscale-url http://localhost:8080
	killall headscale

build:
	GOOS=linux GOARCH=amd64 go build -o dist/truenas-tailscale .

clean:
	rm dist/truenas-tailscale
	rmdir dist
