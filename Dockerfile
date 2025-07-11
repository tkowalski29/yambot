FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o yambot cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

RUN mkdir -p /app/config

COPY --from=builder /app/yambot .

EXPOSE 8080

CMD ["./yambot", "/app/config/config.yml"]