FROM ubuntu:20.04

RUN apt-get update && \
    apt-get install -y \
    curl \
    wget \
    nfs-common \
    mount \
    git \
    build-essential \
    netcat \
    && apt-get clean

RUN wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz -P /tmp && \
    tar -C /usr/local -xvzf /tmp/go1.23.4.linux-amd64.tar.gz && \
    rm /tmp/go1.23.4.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN mkdir -p /mnt/nfs

WORKDIR /app

COPY . ./

RUN go mod tidy

EXPOSE 3000

RUN go build -o distribution-server .

COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

CMD ./distribution-server