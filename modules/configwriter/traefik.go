package configwriter

import (
	"docker-certs/core/configs"
	"docker-certs/core/types"
	"docker-certs/modules/certs"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var mu sync.Mutex

type DynamicConfig struct {
	TLS struct {
		Certificates []struct {
			CertFile string `yaml:"certFile"`
			KeyFile  string `yaml:"keyFile"`
		} `yaml:"certificates"`
	} `yaml:"tls"`
}

func createDynamicYAMLIfNotExists() error {
	certDir := "certs"
	filePath := certDir + "/dynamic.yaml"

	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		if err := os.MkdirAll(certDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creating certs directory: %v", err)
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("Creating dynamic YAML file...")
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file dynamic.yaml: %v", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatalf("There was an error trying to close the dynamic.yaml: %v", err)
			}
		}(file)
	}
	return nil
}

func writeDynamicYAML(certFile, keyFile string) error {
	mu.Lock()
	defer mu.Unlock()

	appConfig := configs.GetConfig()
	filePath := filepath.Join(appConfig.CertsDir, "dynamic.yaml")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := createDynamicYAMLIfNotExists(); err != nil {
			return err
		}
	}
	
	config := DynamicConfig{}

	data, err := os.ReadFile(filePath)

	if err == nil {
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("error unmarshaling existing YAML: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error reading existing YAML file: %v", err)
	}

	certAlreadyExists := false
	for cert := range config.TLS.Certificates {
		if config.TLS.Certificates[cert].CertFile == certFile {
			certAlreadyExists = true
			break
		}
	}

	if !certAlreadyExists {
		config.TLS.Certificates = append(config.TLS.Certificates, struct {
			CertFile string `yaml:"certFile"`
			KeyFile  string `yaml:"keyFile"`
		}{CertFile: certFile, KeyFile: keyFile})
	}

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

func updateTraefikConfig(e types.Event[certs.Event]) {
	log.Println("[config-writer] Certificate created event received")

	if err := writeDynamicYAML(e.Payload.CertFile, e.Payload.KeyFile); err != nil {
		log.Println("[config-writer] Error writing dynamic YAML file:", err)
	}
}
