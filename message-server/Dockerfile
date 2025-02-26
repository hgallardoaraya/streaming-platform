FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o /app/message-server ./...

EXPOSE 9091

CMD ["/app/message-server"]
