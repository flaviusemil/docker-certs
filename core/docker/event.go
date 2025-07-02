package docker

type Event struct {
	ID         string            `json:"id"`
	Action     string            `json:"action"`
	Attributes map[string]string `json:"attributes"`
}
