FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o payment-service ./cmd/payment-service

EXPOSE 8082

CMD ["./payment-service"]
