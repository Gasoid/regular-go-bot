FROM golang:1.18-stretch AS builder
WORKDIR /code
ADD go.mod /code/
ADD go.sum /code/
RUN go mod download
ADD *.go /code/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /code/bot .

FROM alpine:3.6
EXPOSE 8080
WORKDIR /root/
RUN apk --no-cache --update add bash curl less jq openssl
COPY --from=builder /code/bot /root/
HEALTHCHECK --interval=10s --timeout=3s \
  CMD curl -f http://localhost:8080/health || exit 1
CMD exec /root/bot
