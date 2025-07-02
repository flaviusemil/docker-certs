package configwriter

type ConfigUpdatedEvent struct {
	Host     string `json:"host"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}
