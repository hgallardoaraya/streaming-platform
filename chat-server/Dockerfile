FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o /app/chat-server ./...

EXPOSE 9090

CMD ["/app/chat-server"]
