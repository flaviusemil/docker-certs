package mdns

import (
	"docker-certs/core/configs"
	"docker-certs/core/docker"
	"docker-certs/core/eventbus"
	"docker-certs/core/types"
	"docker-certs/modules/certs"
	"github.com/grandcat/zeroconf"
	"log"
	"net"
	"strings"
	"sync"
)

type Module struct {
	mu      sync.Mutex
	servers map[string]*zeroconf.Server
}

func (m *Module) Init() error {
	log.Println("[mdns] Init")
	m.servers = make(map[string]*zeroconf.Server)
	return nil
}

func (m *Module) RegisterEventHandlers() {

	appConfig := configs.GetConfig()

	if appConfig.MDNSPublishing {
		log.Println("[mdns] Registering listeners")

		eventbus.On(types.ContainerStarted, func(e types.Event[docker.Event]) {
			hosts := certs.ExtractHosts(e.Payload.Attributes)
			for _, host := range hosts {
				if err := registerService(m, host); err != nil {
					log.Printf("mDNS register error for host %s: %v", host, err)
				}
			}
		})

		eventbus.On(types.ContainerStopped, func(e types.Event[docker.Event]) {
			hosts := certs.ExtractHosts(e.Payload.Attributes)
			for _, host := range hosts {
				if err := unregisterService(m, host); err != nil {
					log.Printf("mDNS unregister error for host %s: %v", host, err)
				}
			}
		})
	}
}

func registerService(m *Module, host string) error {
	if !strings.HasSuffix(host, ".local") {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[host]; exists {
		return nil
	}

	instance := strings.Split(host, ".")[0]
	service, tld, port := "_https._tcp", "local.", 443
	domain := host[:strings.LastIndex(host, ".")] + "."
	txt := []string{"path=/"}

	srv, err := zeroconf.RegisterProxy(instance, service, tld, port, domain, lanIPs(), txt, nil)
	if err != nil {
		log.Fatal(err)
	}

	m.servers[host] = srv
	log.Printf("Advertised service %s.%s.%s on port %d\n", instance, service, domain, port)
	return nil
}

func unregisterService(m *Module, host string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if srv, exists := m.servers[host]; exists {
		srv.Shutdown()
		delete(m.servers, host)
		log.Printf("[mdns] Unregistered %s", host)
	}

	return nil
}

func lanIPs() (ips []string) {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip, _, _ := net.ParseCIDR(addr.String())
			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil &&
				(ip4[0] == 10 || (ip4[0] == 192 && ip4[1] == 168)) {
				ips = append(ips, ip4.String())
			}
		}
	}
	return
}

func (m *Module) Close() error {
	log.Println("[mdns] Shutting down mDNS services")
	m.mu.Lock()
	defer m.mu.Unlock()

	for h, srv := range m.servers {
		srv.Shutdown()
		delete(m.servers, h)
	}

	return nil
}
