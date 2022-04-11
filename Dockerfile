FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /go/bin/blive

FROM alpine:latest

COPY --from=builder /go/bin/blive /blive
RUN chmod +x /blive

ENV GIN_MODE=release
ENV RESTRICT_GLOBAL=192.168.0.127

EXPOSE 8080

VOLUME /cache

ENTRYPOINT [ "/blive" ]