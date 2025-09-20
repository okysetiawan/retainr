package config

type MinioConfig struct {
	Endpoint   string `mapstructure:"endpoint"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	Secure     bool   `mapstructure:"secure"`
	BucketName string `mapstructure:"bucket_name"`
}
