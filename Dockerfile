FROM golang:1.14 AS build-env
RUN apt-get update -y
RUN apt-get install -y zip
RUN go get -u github.com/go-bindata/go-bindata/...
ADD . /src
RUN cd /src && make linux

FROM alpine:latest
COPY --from=build-env /src/build/linux/x86-64/awslambdaproxy /app/

ENV AWS_ACCESS_KEY_ID=
ENV AWS_SECRET_ACCESS_KEY=
ENV REGIONS=
ENV FREQUENCY=
ENV MEMORY=
ENV SSH_USER=
ENV SSH_PORT=2222
ENV LISTENERS=
ENV DEBUG_PROXY=

WORKDIR /app

RUN addgroup -g 1000 -S ssh \
 && adduser -u 1000 -S ssh -G ssh \
 && apk add --no-cache openssh-server bash ca-certificates \
 && rm -rf /var/cache/apk/*

USER ssh

RUN mkdir ${HOME}/.ssh

EXPOSE 2222
EXPOSE 8080

COPY docker/sshd_config /etc/ssh/sshd_config
COPY docker/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]