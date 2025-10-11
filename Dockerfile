FROM golang:1.25.1-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/port-service
RUN go build -o app main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/cmd/port-service/app .
EXPOSE 8080
CMD ["./app"]
