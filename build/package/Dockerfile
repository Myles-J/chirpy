FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o chirpy ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/chirpy .
COPY --from=builder /app/configs/sql /app/configs/sql

EXPOSE 8080
CMD ["./chirpy"] 