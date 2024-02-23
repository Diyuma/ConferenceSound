FROM golang:latest

WORKDIR /app

COPY . .
# TODO later copy not all

RUN go build -o rest_server

EXPOSE 8086

CMD ["./rest_server"]