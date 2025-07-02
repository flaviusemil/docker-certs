package certs

type Event struct {
	Host     string `json:"host"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}
