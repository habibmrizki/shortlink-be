FROM golang:1.27.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd

FROM alpine:3.24

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server ./server

EXPOSE 2901

CMD ["/app/server"]
