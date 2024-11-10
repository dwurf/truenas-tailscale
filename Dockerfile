FROM golang:1.23-alpine AS builder
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o truenas-tailscale .

FROM alpine

COPY --from=builder /usr/src/app/truenas-tailscale /
VOLUME /root/.config/truenas-tailscale/

CMD ["/truenas-tailscale"]
