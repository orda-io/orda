# Ortoo
[Ortoo (or Yam)](https://en.wikipedia.org/wiki/Yam_(route) was a mongolian messenger system. Ortoo project is a multi-device data synchronization platform based on MongoDB (which could be other document databases). Ortoo is based on CRDT(conflict-free data types), which enables operation-based syncronization.  


## Getting started

### Working envirnment (Maybe work on less versions of them)
 - go 1.13.5
 - docker 18.09.2 (for running MongoDB)
 - [MongoDB latest](https://hub.docker.com/_/mongo)
 - MQTT: [eclipse mosquitto latest](https://hub.docker.com/_/eclipse-mosquitto) 
 - docker-compose
 - gogo/protobuf (how to install: http://google.github.io/proto-lens/installing-protoc.html)
 
### Install
 ```bash
 # git clone https://github.com/knowhunger/ortoo.git
 # cd ortoo 
 # make docker-up
 # make protoc-gen
 ```

## For developers 

### Reading first 
  - Overall architecture of Ortoo. 
  - Coding conventions 
