# 🐎🎪Orda

----

```

    .::::                 .::          
  .::    .::              .::          
.::        .::.: .:::     .::   .::    
.::        .:: .::    .:: .:: .::  .:: 
.::        .:: .::   .:   .::.::   .:: 
  .::     .::  .::   .:   .::.::   .:: 
    .::::     .:::    .:: .::  .:: .:::
                                          
```

[Orda (or ordu, ordo, or ordon)](https://en.wikipedia.org/wiki/Orda_(organization)) means the Mongolian mobile tent or
place tent. As stated in [Civilization WIKI](https://civilization.fandom.com/wiki/Ordu_(Civ6)), Orda (or Ordu) was the
center of the tribe for the nomadic Mongolians. Orda project is a multi-device data synchronization platform based on
[MongoDB](https://www.mongodb.com/) (which could be other document databases such
as [CouchBase](https://www.couchbase.com/)). Orda is based on CRDT(conflict-free data types), which enables
operation-based synchronization.

This project is mainly based on these two papers.

- Hyun-Gul Roh et
  al. [Replicated abstract data types: Building blocks for collaborative applications](https://www.sciencedirect.com/science/article/abs/pii/S0743731510002716)
  , JPDC, March, 2011.
- Hyun-Gul Roh et
  al. [Kaleido: Implementing a Novel Data System for Multi-Device Synchronization](https://ieeexplore.ieee.org/document/7962464)
  , IEEE MDM, July, 2017.

## Concepts

- Using Orda SDKs, you can allow multiple users to synchronize any document in a collection of DocumentDB.

!<img src="https://user-images.githubusercontent.com/3905310/128593526-747bb040-6952-4204-b99a-80ebd6c50170.png" width="700"/>

- Any field can be added, removed and updated using JSON operations.
- For example, [Orda-JSONEditor](https://github.com/orda-io/orda-jsoneditor) implemented
  with [Orda-js](https://github.com/orda-io/orda-js) allows multiple users concurrently to edit any document in a
  collection of MongoDB as the following picture.

![ezgif com-gif-maker](https://user-images.githubusercontent.com/3905310/128254096-cf0a9238-2337-4153-8a5d-a91db78e0607.gif)

- You can see the full video of the example usage of Orda JSONEditor At
  YouTube [here](https://www.youtube.com/watch?v=t_R47AWMv6s).

- The documents in the MongoDB collection can be used for the analytics or read-only workloads. It is not recommended
  modifying them directly.

## Working environment

- go 1.18
- docker latest, docker-compose (for building and deploying local orda-server)
- protobuf / grpc / grpc-gateway
- [MongoDB latest](https://hub.docker.com/_/mongo)
- MQTT: [EQM](https://www.emqx.io/)
- Redis latest
- Maybe work on lower versions of them

## Getting started

### Install

 ```bash
 $ git clone https://github.com/orda-io/orda.git
 $ cd orda
 $ make init # generate builder container
 $ make protoc-gen # generate protobuf codes
 ```

### Run Orda Server

- You can run orda server with docker-compose.

```bash
 $ make build-docker-server
 $ make docker-up
 $ docker ps 
 ```

- Port mapping in docker-compose
  * orda-server
    - 19061: for gRPC protocol
    - 19861: for HTTP REST control ([swagger](http://localhost:19861/swagger))
  * orda-mongodb
    - 27017
  * orda-emqx
    - 18181(host):1883(internal) : for MQTT
    - 18881(host):8083(internal) : for websocket / HTTP
    - 18081(host):18083(internal) : for [dashboard](http://localhost:18081)
  * orda-redis
    - 16379(host):6379(internal) : for client

### Make a collection

- To make Orda clients connect to Orda server, a `collection` should be prepared in MongoDB.
- The `collection` is going to create a real collection of MongoDB.
- It can be created by restful API: `curl -X GET http://<SERVER_ADDR>/api/v1/collections/<COLLECTION_NAME>`
- The collections can be created and reset by the [swagger](http://localhost:19861/swagger).

```bash
 $ curl -s -X PUT http://localhost:19861/api/v1/collections/hello_world
Created collection 'hello_world(1)'
```

### Use Orda client SDK

#### Make a client

- An Orda client manages the connection with the Orda server and synchronization of Orda datatypes.
- An Orda client participates in a collection of MongoDB, which means that the snapshot of any created datatype is
  written to the collection of MongoDB.
- The orda client package should be imported as follows.

```shell
$ cd YOUR_PORJECT_PACKAGE
$ go get -u github.com/orda-io/orda/client
```

- To create client, you

```go
clientConf := &orda.ClientConfig{
ServerAddr:       "localhost:19061", // Orda Server address.
NotificationAddr: "localhost:11883", // notification server address.
CollectionName:   "hello_world", // the collection name of MongoDB which the client participates in.
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
Counter := client.CreateCounter("key", orda.NewHandlers(...)

// SubscribeXXXX() is used to subscribe an existing datatype. 
// If there exists no datatype for the key, it returns an error via error handler
Counter := client.SubscribeCounter("key", orda.NewHandlers(...)

// SubscribeOrCreateXXXX() is used to subscribe an existing datatype. 
// If no datatype exists for the key, a new datatype is created. 
// It is recommended to use this method to shorten code lines.
Counter := client.SubscribeOrCreateCounter("key", orda.NewHandlers(...)

// Client should sync with Orda server.
if err := client.Sync(); err !=nil {
panic(err)
}
```

## Contribute to Orda Project

----
We always welcome your participation. Please see [contributing doc](CONTRIBUTING.md).
If you're also interested in writing academic papers with our orda implementations such as JSON CRDTs, feel free contact
us:
<img style="margin-bottom:-5px" height="20" src=".\assets\email-image.png"/>

## License

----
Orda is licensed under Apache 2.0 License that can be found in the LICENSE file. 
