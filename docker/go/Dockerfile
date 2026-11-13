FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY api/go.* ./
RUN go mod download

COPY api/ .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/app /app

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app"]
