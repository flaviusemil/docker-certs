package certs

import (
	"crypto/x509"
	"docker-certs/core/configs"
	"docker-certs/core/docker"
	"docker-certs/core/eventbus"
	"docker-certs/core/types"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Module struct {
	listeners []eventbus.ListenerHandle
}

func (m *Module) Init() error { return nil }

func (m *Module) RegisterEventHandlers() {
	eventbus.On(types.ContainerStarted, func(e types.Event[docker.Event]) {
		hosts := ExtractHosts(e.Payload.Attributes)

		for _, h := range hosts {
			if err := createCertificateIfNeeded(h); err != nil {
				log.Printf("[certs] Error creating certificate for host %s: %s", h, err)
			}
		}
	})
}

func ExtractHosts(actorMap map[string]string) []string {
	var hosts []string
	for key, value := range actorMap {
		if strings.HasPrefix(key, "traefik.http.routers.") && strings.HasSuffix(key, ".rule") {
			labeledHosts := strings.Split(value, "||")
			for _, host := range labeledHosts {
				host = extractHost(host)
				if host != "" && !strings.ContainsAny(host, "{}") {
					hosts = append(hosts, host)
				}
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
	appConfig := configs.GetConfig()
	certDir := appConfig.CertsDir
	certFile := filepath.Join(certDir, host+".pem")
	keyFile := filepath.Join(certDir, host+"-key.pem")

	if certExistsAndValid(certFile) {
		log.Printf("[certs] Certificate for host %s already exists and is valid.\n", host)
		return nil
	}

	if err := createCertificate(host, certFile, keyFile); err != nil {
		return err
	}

	return nil
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
	rootCertFile := filepath.Join(getCARoot(), "rootCA.pem")
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

func getCARoot() string {
	if root := os.Getenv("CAROOT"); root != "" {
		return root
	}
	out, err := exec.Command("mkcert", "-CAROOT").Output()
	if err != nil {
		log.Println("Failed to get CAROOT via mkcert:", err)
		return "."
	}
	return strings.TrimSpace(string(out))
}

func createCertificate(host, certFile, keyfile string) error {
	log.Printf("[certs] Creating certificate for host %s", host)

	appCfg := configs.GetConfig()

	certDir := appCfg.CertsDir
	if err := os.MkdirAll(certDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create certs directory: %v", err)
	}

	cmd := exec.Command("mkcert", "-cert-file", certFile, "-key-file", keyfile, host)

	if appCfg.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error creating certificate for host %s: %v", host, err)
	}

	eventbus.Publish(types.Event[Event]{
		Type: types.CertCreated,
		Payload: Event{
			Host:     host,
			CertFile: filepath.ToSlash(certFile),
			KeyFile:  filepath.ToSlash(keyfile),
		},
	})

	return nil
}
