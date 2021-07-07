# Orda
```
    .::::                 .::          
  .::    .::              .::          
.::        .::.: .:::     .::   .::    
.::        .:: .::    .:: .:: .::  .:: 
.::        .:: .::   .:   .::.::   .:: 
  .::     .::  .::   .:   .::.::   .:: 
    .::::     .:::    .:: .::  .:: .:::                                       
```

[Orda (or ordu, ordo, or ordon)](https://en.wikipedia.org/wiki/Orda_(organization)) means the Mongolian mobile tent or place tent.
As stated in [Civilization WIKI](https://civilization.fandom.com/wiki/Ordu_(Civ6)), Orda(or Ordu) was the center of the tribe for the nomadic Mongolians.
Orda project is a multi-device data synchronization platform based on MongoDB (which could be other document databases). 
Orda is based on CRDT(conflict-free data types), which enables operation-based synchronization.  


## Getting started

### Working envirnment (Maybe work on less versions of them)

- go 1.16
- docker latest, docker-compose (for deploying local orda-server)
- [MongoDB latest](https://hub.docker.com/_/mongo)
- MQTT: [EQM](https://www.emqx.io/)
- gogo/protobuf (how to install: http://google.github.io/proto-lens/installing-protoc.html)


## How to use Orda
 
### Running local Orda server with docker-compose.
 ```bash
 # git clone https://github.com/orda-io/orda.git
 # cd orda
 # make protoc-gen
 # make build-local-docker-server
 # make docker-up
 # make server
 ```

- Port mapping in docker-compose
  * orda-server
    - 19061: for gRPC protocol
    - 19861: for HTTP REST control
  * orda-mongodb
    - 27017
  * orda-emqx
    - 18181(host):1883(internal) : for MQTT
    - 18881(host):8083(internal) : for websocket / HTTP
    - 18081(host):18083(internal) : for [dashboard](http://localhost:18081))
  * orda-envoy-grpc-proxy
    - 29065: for grpc-web (websocket of orda-emqx)

### Use Orda client SDK

#### Make a collection
 - To make Orda clients connect to Orda server, a `collection` should be prepared. The `collection` corresponds to the real collection of MongoDB. It can be created by restful API: `curl -X GET http://<SERVER_ADDR>/collections/<COLLECTION_NAME>`
```bash
 $ curl -s -X PUT http://localhost:19861/collections/hello_world
Created collection 'hello_world(1)'
```
#### Make a client
 - An Orda client manages the connection with the Orda server and synchronization of Orda datatypes.   
 - An Orda client participates in a collection of MongoDB, which means that the snapshot of any created datatype is written to the collection of MongoDB.   
```go
clientConf := &orda.ClientConfig{
    ServerAddr:       "localhost:19061",         // Orda Server address.
    NotificationAddr: "localhost:11883",          // notification server address.
    CollectionName:   "hello_world",             // the collection name of MongoDB which the client participates in.
    SyncType:         model.SyncType_REALTIME, // syncType that is notified in real-time from notification server.
}

client1 := orda.NewClient(clientConf, "client1") // create a client with alias "client1".
if err := client1.Connect(); err != nil {         // connect to Orda server
    _ = fmt.Errorf("fail client to connect an Orda server: %v", err.Error())
    return
}

defer func() {
    if err := client1.Close(); err != nil { // close the client
        _ = fmt.Errorf("fail to close client: %v", err.Error())
    }
}()
```
### Use Datatypes

#### Creation and Subscription of Datatypes
 - For using datatypes, a client should create or subscribe datatypes. 
 - A datatype has a key which is used as a document ID; i.e., `_id' of MongoDB document.
 - There are three ways to use a datatype. For a type of datatype `XXXX`, `XXXX` cam be one of IntCounter, HashMap, List... 
```go
// CreateXXXX() is used to create a new datatype. 
// If there already exists a datatype for the key, it returns an error via error handler.
intCounter := client.CreateCounter("key", orda.NewHandlers(...)

// SubscribeXXXX() is used to subscribe an existing datatype. 
// If there exists no datatype for the key, it returns an error via error handler
intCounter := client.SubscribeCounter("key", orda.NewHandlers(...)

// SubscribeOrCreateXXXX() is used to subscribe an existing datatype. 
// If no datatype exists for the key, a new datatype is created. 
// It is recommended to use this method to shorten code lines.
intCounter := client.SubscribeOrCreateCounter("key", orda.NewHandlers(...)

// Client should sync with Orda server.
if err:= client.Sync(); err !=nil {
    panic(err)
}
```

#### Handlers

#### Counter

#### Map

#### List

#### Transaction

