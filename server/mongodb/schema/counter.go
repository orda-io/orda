package schema

type CounterDoc struct {
	ID  string `bson:"_id"`
	Num uint32 `bson:"num"`
}
