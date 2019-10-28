package schema

type CounterDoc struct {
	ID  string `bson:"_id"`
	Num uint32 `bson:"num"`
}

var CounterDocFields = struct {
	ID  string
	Num string
}{
	ID:  "_id",
	Num: "num",
}
