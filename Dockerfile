FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY . .

# private go module
# 僅限 v0.1.3 版本, v0.1.4 版本後移除

ARG ACCESS_TOKEN

RUN apk add git

RUN git config --global url.https://$ACCESS_TOKEN:x-oauth-basic@github.com/eric2788.insteadOf https://github.com/eric2788

# ================

RUN go mod download

RUN go build -o /go/bin/blive

FROM alpine:latest

COPY --from=builder /go/bin/blive /blive
RUN chmod +x /blive

ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT [ "/blive" ]