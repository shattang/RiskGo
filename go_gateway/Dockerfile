FROM golang:alpine AS builder

RUN apk add --no-cache protoc protobuf-dev

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go_gateway/go.mod go_gateway/go.sum ./go_gateway/
WORKDIR /app/go_gateway
RUN go mod download

WORKDIR /app
COPY protos/ protos/
COPY go_gateway/ go_gateway/

# Generate gRPC code again in container to be sure
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN protoc --proto_path=protos --go_out=go_gateway/proto --go_opt=paths=source_relative --go-grpc_out=go_gateway/proto --go-grpc_opt=paths=source_relative protos/risk_engine.proto

WORKDIR /app/go_gateway
RUN go mod tidy
RUN go build -o gateway main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/go_gateway/gateway .
COPY LICENSE NOTICE /usr/local/share/doc/riskgo/
EXPOSE 3000
CMD ["./gateway"]
