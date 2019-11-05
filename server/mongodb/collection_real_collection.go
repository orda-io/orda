package mongodb

//CollectionSnapshots is used for manipulating snapshot of datatypes.
type CollectionSnapshots struct {
	*MongoCollections
	name string
}
