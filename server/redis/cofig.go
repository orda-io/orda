package redis

type Config struct {
	Addrs    []string `json:"Addrs"`
	Username string   `json:"Username"`
	Password string   `json:"Password"`
}
