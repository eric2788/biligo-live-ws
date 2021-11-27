FROM golang:latest

ENV GIN_MODE=release

WORKDIR /app

RUN cd /app

COPY . .

RUN go mod download

RUN go build -o /program

EXPOSE 8080

CMD [ "/program" ]