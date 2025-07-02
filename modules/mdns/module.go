package mdns

import (
	"docker-certs/core/eventbus"
	"docker-certs/core/types"
	"docker-certs/modules/configwriter"
	"github.com/grandcat/zeroconf"
	"log"
	"net"
	"strings"
)

type Module struct {
	servers []*zeroconf.Server
}

func (m *Module) Init() error {
	log.Println("[mdns] Init")

	return nil
}

func (m *Module) RegisterListeners(bus *eventbus.EventBus) {
	log.Println("[mdns] Registering listeners")

	eventbus.On(types.ConfigUpdated, func(e types.Event[configwriter.ConfigUpdatedEvent]) {
		if err := registerService(m, e.Payload.Host); err != nil {
			log.Printf("mDNS register error for host %s: %v", e.Payload.Host, err)
		}
	})
}

//bus.RegisterListener("stop", func(e types.Event) {
//	for _, host := range e.Hosts {
//		unregisterService(host)
//	}
//})

func registerService(m *Module, host string) error {
	if !strings.HasSuffix(host, ".local") {
		return nil
	}

	index := strings.Index(host, ".")
	instance := host[:index]
	service := "_https._tcp"
	tld := "local."
	port := 443
	domain := host[:strings.LastIndex(host, ".")] + "."
	//ips := []string{"10.0.20.2"}
	txt := []string{"path=/"}

	srv, err := zeroconf.RegisterProxy(instance, service, tld, port, domain, lanIPs(), txt, nil)
	if err != nil {
		log.Fatal(err)
	}

	m.servers = append(m.servers, srv)

	log.Printf("Advertised service %s.%s.%s on port %d\n", instance, service, domain, port)
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
			// keep only RFC‑1918 private LAN ranges:
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

	for _, srv := range m.servers {
		srv.Shutdown()
	}

	return nil
}
