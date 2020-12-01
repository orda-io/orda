FROM ubuntu:20.10
MAINTAINER Hyun-Gul Roh<knowhunger@gmail.com>

RUN apt-get update
RUN mkdir /app /config
COPY ./tmp/server /app/ortoo_server


WORKDIR /app

CMD ["/app/ortoo_server", "--conf", "/config/local-config.json"]

EXPOSE 19061
EXPOSE 19861