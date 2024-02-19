FROM golang:latest

WORKDIR /app

COPY . .
# TODO later copy not all

RUN go build -C server/cmd -o server

EXPOSE 9090

CMD ["./server/cmd/server", "-sredisaddr", "172.18.0.5:6379"]