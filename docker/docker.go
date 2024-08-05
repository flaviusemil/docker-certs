package docker

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/grandcat/zeroconf"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ListenToDockerEvents() {
	if err := createDynamicYAMLIfNotExists(); err != nil {
		log.Fatalf("There was an error trying to create dynamic YAML files: %v", err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error createing Docker client: %v", err)
	}

	messages, errs := cli.Events(context.Background(), events.ListOptions{})
	fmt.Println("Listening for Docker container events...")

	for {
		select {
		case msg := <-messages:
			if msg.Type == "container" && msg.Action == "start" {
				event := parseDockerEvent(msg)
				fmt.Printf("Container Name: %s - Action: %s - Hosts: %v\n", event.ContainerID, event.Action, event.Hosts)
				for _, host := range event.Hosts {
					err := createCertificateIfNeeded(host)

					if err != nil {
						log.Printf("Error creating certificate for host %s: %v", host, err)
					}
				}
			}
		case err := <-errs:
			log.Fatalf("Error listening for events: %v", err)
		}
	}
}

func createDynamicYAMLIfNotExists() error {
	certDir := "certs"

	if err := os.MkdirAll(certDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating certs directory: %v", err)
	}

	file, err := os.Create("certs/dynamic.yaml")
	if err != nil {
		return fmt.Errorf("error creating file dynamic.yaml: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("There was an error trying to close the dynamic.yaml: %v", err)
		}
	}(file)
	return nil
}

func parseDockerEvent(msg events.Message) Event {
	actorMap := make(map[string]string)
	for k, v := range msg.Actor.Attributes {
		actorMap[k] = v
	}

	hosts := extractHosts(actorMap)

	return Event{
		ContainerID: msg.Actor.ID,
		Action:      msg.Action,
		Actor:       actorMap,
		Hosts:       hosts,
	}
}

func extractHosts(actorMap map[string]string) []string {
	var hosts []string
	for key, value := range actorMap {
		if strings.HasPrefix(key, "traefik.http.routers.") && strings.HasSuffix(key, ".rule") {
			host := extractHost(value)
			if host != "" && !strings.ContainsAny(host, "{}") {
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}

func extractHost(rule string) string {
	start := strings.Index(rule, "`")
	end := strings.LastIndex(rule, "`")

	if start != -1 && end != -1 && start < end {
		return rule[start+1 : end]
	}
	return ""
}

func createCertificateIfNeeded(host string) error {
	certDir := "certs"
	certFile := filepath.Join(certDir, host+".pem")
	keyFile := filepath.Join(certDir, host+"-key.pem")

	if err := registerService(host); err != nil {
		return err
	}

	if certExistsAndValid(certFile) {
		fmt.Printf("Certificate for host %s already exists and is valid.\n", host)
		return nil
	}

	if err := createCertificate(host, certFile, keyFile); err != nil {
		return err
	}

	return writeDynamicYAML(certFile, keyFile)
}

func certExistsAndValid(certFile string) bool {
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return false
	}

	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		log.Printf("Error reading certificate file %s: %v", certFile, err)
		return false
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		log.Printf("Failed to decode PEM block containing the certificate")
		return false
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Printf("Error parsing certificate: %v", err)
		return false
	}

	rootCertPool := x509.NewCertPool()
	rootCertFile := filepath.Join(os.Getenv("CAROOT"), "rootCA.pem")
	rootCertPEM, err := os.ReadFile(rootCertFile)
	if err != nil {
		log.Printf("Error reading root certificate file %s: %v", rootCertFile, err)
		return false
	}
	rootCertPool.AppendCertsFromPEM(rootCertPEM)

	opts := x509.VerifyOptions{
		Roots: rootCertPool,
	}

	if _, err := cert.Verify(opts); err != nil {
		log.Printf("Certificate verification failed: %v", err)
		return false
	}

	return true
}

func createCertificate(host, certFile, keyfile string) error {
	cmd := exec.Command("mkcert", "-cert-file", certFile, "-key-file", keyfile, host)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error creating certificate for host %s: %v", host, err)
	}
	return nil
}

func writeDynamicYAML(certFile, keyFile string) error {
	config := DynamicConfig{}

	data, err := os.ReadFile("certs/dynamic.yaml")

	if err == nil {
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("error unmarshaling existing YAML: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error reading existing YAML file: %v", err)
	}

	config.TLS.Certificates = append(config.TLS.Certificates, struct {
		CertFile string `yaml:"certFile"`
		KeyFile  string `yaml:"keyFile"`
	}{CertFile: certFile, KeyFile: keyFile})

	data, err = yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("error marshaling YAML: %v", err)
	}

	err = os.WriteFile("certs/dynamic.yaml", data, 0644)
	if err != nil {
		return fmt.Errorf("error writing YAML file: %v", err)
	}
	log.Println("YAML file created successfully")
	return nil
}

func registerService(host string) error {
	index := strings.Index(host, ".")
	instance := host[:index]
	service := "_http._tcp"
	domain := "local."
	port := 443
	txt := []string{"path=/"}

	server, err := zeroconf.Register(instance, service, domain, port, txt, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Shutdown()
	log.Printf("Advertised service %s.%s.%s on port %d\n", instance, service, domain, port)
	return nil
}
