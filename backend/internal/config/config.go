package config

type Config struct {
	Name string `yaml:"Name"`
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`

	Auth struct {
		AccessSecret string `yaml:"AccessSecret"`
		AccessExpire int64  `yaml:"AccessExpire"`
	} `yaml:"Auth"`

	Database struct {
		DataSource string `yaml:"DataSource"`
	} `yaml:"Database"`

	VideoDir string `yaml:"VideoDir"`

	Log struct {
		Mode  string `yaml:"Mode"`
		Level string `yaml:"Level"`
	} `yaml:"Log"`
}
