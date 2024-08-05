package docker

type DynamicConfig struct {
	TLS struct {
		Certificates []struct {
			CertFile string `yaml:"certFile"`
			KeyFile  string `yaml:"keyFile"`
		} `yaml:"certificates"`
	} `yaml:"tls"`
}
