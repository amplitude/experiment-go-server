FROM golang:1.18-buster

#ENV PATH=$PATH:/usr/local/go/bin
#ENV GOROOT=/usr/local/go
#
#RUN apt-get update && \
#    apt-get install -y wget git make gcc g++ && \
#    wget https://dl.google.com/go/go1.16.4.linux-amd64.tar.gz && \
#    tar -C /usr/local -xzf go1.16.4.linux-amd64.tar.gz && \
#    rm -rf /go1.16.4.linux-amd64.tar.gz

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and
# only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN make