FROM ubuntu:20.10
MAINTAINER Hyun-Gul Roh<knowhunger@gmail.com>

RUN apt-get update
RUN mkdir /app /config
COPY ./tmp/server /app/orda_server


WORKDIR /app

CMD ["/app/orda_server", "--conf", "/config/docker-config.json"]

EXPOSE 19061
EXPOSE 19861