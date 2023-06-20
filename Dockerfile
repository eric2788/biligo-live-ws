FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .

ARG VERSION

RUN go mod download

RUN go build -ldflags "-X main.version=$VERSION" -o /go/bin/blive

FROM alpine:latest

COPY --from=builder /go/bin/blive /blive
RUN chmod +x /blive

ENV GIN_MODE=release
ENV RESTRICT_GLOBAL=changeme
ENV NO_LISTENING_LOG=true
ENV RESET_LOW_LATENCY=false

EXPOSE 8080

VOLUME /cache

ENTRYPOINT [ "/blive" ]