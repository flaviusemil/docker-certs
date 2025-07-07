FROM golang:latest AS builder

ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o docker-certs .

FROM alpine:latest

ARG TARGETARCH

RUN apk update
RUN apk --no-cache add ca-certificates curl \
    && curl -Lo /usr/local/bin/mkcert https://dl.filippo.io/mkcert/latest?for=linux/amd64 \
    && chmod +x /usr/local/bin/mkcert

ENV CAROOT=/root/rootCerts

WORKDIR /app

COPY --from=builder /app/docker-certs .
CMD ["./docker-certs"]
