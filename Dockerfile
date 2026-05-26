FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o weather-service .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/weather-service .
EXPOSE 8080
ENTRYPOINT ["./weather-service"]
