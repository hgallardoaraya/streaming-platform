FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY ./rtmp-server /rtmp-server

EXPOSE 1935

CMD ["./rtmp-server"]
