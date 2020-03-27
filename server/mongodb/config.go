package mongodb

// Config is a configuration for MongoDB
type Config struct {
	Host    string `json:"MongoHost"`
	OrtooDB string `json:"MongoOrtoo"`
}

// NewTestMongoDBConfig creates a new MongoDBConfig for Test
func NewTestMongoDBConfig(dbName string) *Config {
	return &Config{
		Host:    "mongodb://root:ortoo-test@localhost:27017",
		OrtooDB: dbName,
	}
}
