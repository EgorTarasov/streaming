package kafka

type Config struct {
	Brokers []Broker `yaml:"brokers"`
	Topics  []Topic  `yaml:"topics"`
}

type Broker struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Topic struct {
	Name       string `yaml:"name"`
	Partitions int32  `yaml:"partitions"`
}
