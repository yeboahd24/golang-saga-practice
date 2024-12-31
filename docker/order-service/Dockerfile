FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o order-service ./cmd/order-service

EXPOSE 8080

CMD ["./order-service"]
