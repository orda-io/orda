package schema

// CounterDoc defines a document for counter stored in MongoDB. Counter is used to assign number to a collection
type CounterDoc struct {
	ID  string `bson:"_id"`
	Num int32  `bson:"num"`
}

// CounterDocFields defines the fields of CounterDoc
var CounterDocFields = struct {
	ID  string
	Num string
}{
	ID:  "_id",
	Num: "num",
}
