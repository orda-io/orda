package mongodb

// Config is a configuration for MongoDB
type Config struct {
	Host   string `json:"MongoHost"`
	OrdaDB string `json:"OrdaDB"`
}

// NewTestMongoDBConfig creates a new MongoDBConfig for Test
func NewTestMongoDBConfig(dbName string) *Config {
	return &Config{
		Host:   "mongodb://root:orda-test@localhost:27017",
		OrdaDB: dbName,
	}
}
