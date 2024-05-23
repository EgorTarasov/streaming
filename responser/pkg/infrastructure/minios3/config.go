package minios3

type Config struct {
	Host      string   `yaml:"host"`
	Port      int      `yaml:"port"`
	AccessKey string   `yaml:"access_key"`
	SecretKey string   `yaml:"secret_key"`
	Ssl       bool     `yaml:"ssl"`
	Buckets   []Bucket `yaml:"buckets"`
}

type Bucket struct {
	Name   string `yaml:"name"`
	Region string `yaml:"region"`
	Lock   bool   `yaml:"lock"`
}
