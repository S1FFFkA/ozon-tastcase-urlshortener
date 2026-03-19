FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/urlshortener ./cmd

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app
USER app

WORKDIR /home/app

COPY --from=builder /bin/urlshortener /usr/local/bin/urlshortener

EXPOSE 8080

ENTRYPOINT ["urlshortener"]
