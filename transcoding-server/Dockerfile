FROM golang as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o transcoding-server .

FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y nfs-kernel-server rpcbind ffmpeg && \
    mkdir -p /srv/nfs && \
    chown nobody:nogroup /srv/nfs && \
    mkdir -p /run/sendsigs.omit.d && \
    chown root:root /run/sendsigs.omit.d && \
    echo "/srv/nfs *(rw,sync,no_subtree_check)" > /etc/exports && \
    apt-get clean

EXPOSE 2049

WORKDIR /app

COPY --from=builder /app/transcoding-server .

EXPOSE 8080

CMD ["bash", "-c", "service rpcbind start && service nfs-kernel-server start && ./transcoding-server && tail -f /dev/null"]
