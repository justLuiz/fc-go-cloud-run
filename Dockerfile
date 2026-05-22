FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o weather-service .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/weather-service .
EXPOSE 8080
ENTRYPOINT ["./weather-service"]
