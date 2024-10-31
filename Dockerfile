FROM golang:1.23.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /home/aka/Templates/simple_socket

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o pineywss cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /home/aka/Templates/simple_socket

COPY --from=builder /home/aka/Templates/simple_socket .

EXPOSE 8081 8083

CMD ["./pineywss"]
