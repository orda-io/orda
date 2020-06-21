# Ortoo
````
     _/_/                _/                         
  _/    _/  _/  _/_/  _/_/_/_/    _/_/      _/_/    
 _/    _/  _/_/        _/      _/    _/  _/    _/   
_/    _/  _/          _/      _/    _/  _/    _/    
 _/_/    _/            _/_/    _/_/      _/_/
````

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
 # make server
 ```

## How to use Ortoo

### Run Ortoo Server
```bash
 $ make run-local-server 
or 
 $ make server 
 $ ./build/server --conf ./examples/local-config.json 
```

### Use Ortoo client SDK

#### Make a collection
 - To make Ortoo clients connect to Ortoo server, a `collection` should be prepared. The `collection` corresponds to the real collection of MongoDB. It can be created by restful API: `curl -X GET http://<SERVER_ADDR>/collections/<COLLECTION_NAME>`
```bash
 $ curl -s -X GET http://localhost:19861/collections/hello_world
Created collection 'hello_world(1)'
```
#### Make a client
 - An Ortoo client manages the connection with the Ortoo server and synchronization of Ortoo datatypes.   
 - An Ortoo client participates in a collection of MongoDB, which means that the snapshot of any created datatype is written to the collection of MongoDB.   
```go
clientConf := &ortoo.ClientConfig{
    ServerAddr:       "localhost:19061",         // Ortoo Server address.
    NotificationAddr: "localhost:1883",          // notification server address.
    CollectionName:   "hello_world",             // the collection name of MongoDB which the client participates in.
    SyncType:         model.SyncType_NOTIFIABLE, // syncType that is notified in real-time from notification server.
}

client1 := ortoo.NewClient(clientConf, "client1") // create a client with alias "client1".
if err := client1.Connect(); err != nil {         // connect to Ortoo server
    _ = fmt.Errorf("fail client to connect an Ortoo server: %v", err.Error())
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
intCounter := client.CreateIntCounter("key", ortoo.NewHandlers(...)

// SubscribeXXXX() is used to subscribe an existing datatype. 
// If there exists no datatype for the key, it returns an error via error handler
intCounter := client.SubscribeIntCounter("key", ortoo.NewHandlers(...)

// SubscribeOrCreateXXXX() is used to subscribe an existing datatype. 
// If no datatype exists for the key, a new datatype is created. 
// It is recommended to use this method to shorten code lines.
intCounter := client.SubscribeOrCreateIntCounter("key", ortoo.NewHandlers(...)

// Client should sync with Ortoo server.
if err:= client.Sync(); err !=nil {
    panic(err)
}
```
#### Handlers   
#### IntCounter
#### HashMap
#### List
#### Transaction

