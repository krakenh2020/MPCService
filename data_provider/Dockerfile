FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update --fix-missing && apt-get upgrade -y
RUN apt-get install -y wget

WORKDIR /

## Install golang
RUN wget --no-check-certificate https://dl.google.com/go/go1.17.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

## Install the dependencies
ADD . / /root/go/src/github.com/krakenh2020/MPCService/
WORKDIR /root/go/src/github.com/krakenh2020/MPCService
RUN go install
ENV PATH="/root/go/bin:${PATH}"
ENV GOPATH="/root/go"

# Start the MPC server
ENTRYPOINT ["MPCService"]
