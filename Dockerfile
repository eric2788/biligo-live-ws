FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /go/bin/blive

FROM alpine:latest

COPY --from=builder /go/bin/blive /blive
RUN chmod +x /blive

ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT [ "/blive" ]