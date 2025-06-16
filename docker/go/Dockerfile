FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY api/go.* ./
RUN go mod download

COPY api/ .

RUN go build -o app .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
