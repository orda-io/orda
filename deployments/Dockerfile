FROM ubuntu:latest
MAINTAINER Hyun-Gul Roh<knowhunger@gmail.com>

RUN apt-get update
RUN apt-get -y upgrade
RUN mkdir /app /config /app/resources
COPY ./tmp/resources /app/resources
COPY ./tmp/server /app/orda_server

WORKDIR /app

CMD ["/app/orda_server", "--conf", "/config/docker-config.json"]

EXPOSE 19061
EXPOSE 19861