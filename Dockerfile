# Build stage
FROM golang:1.23 AS builder

ENV GOPROXY=direct
ENV GODEBUG=netdns=go

# ENV GOPROXY=direct
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .

EXPOSE 8080
CMD [ "/app/main" ]