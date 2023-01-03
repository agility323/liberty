FROM debian:buster
MAINTAINER Thomas Wade <wutian323@126.com>

ENV LOG_LEVEL 1
ENV HOST 101
ENV CLIENT_SERVER_ADDR 127.0.0.1:4001
ENV ENTRANCE_ADDR 127.0.0.1:4001
ENV CONNECT_SERVER_HANDLER {\"service\": \"login_service\", \"method\": \"connect_server\"}
ENV ETCD [\"http://127.0.0.1:2379\", \"http://127.0.0.1:2479\", \"http://127.0.0.1:2579\"]
ENV PRIVATE_RSA_KEY ./rsa_key

WORKDIR /lserver

COPY ./gate ./
COPY ./rsa_key ./
COPY ./run.sh ./

ENTRYPOINT ["bash", "run.sh"]
