package schema

import "go.mongodb.org/mongo-driver/mongo"

//MongoDBDoc is an interface for documents stored in MongoDB
type MongoDBDoc interface {
	GetIndexModel() []mongo.IndexModel
}
