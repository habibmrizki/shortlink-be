FROM golang:1.27.2-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd

FROM alpine:3.21

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server ./server

EXPOSE 2901

CMD ["/app/server"]
