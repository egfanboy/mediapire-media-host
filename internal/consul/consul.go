package consul

import (
	"fmt"
	"net"
	"strconv"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

var consulClient *api.Client

func NewConsulClient() error {
	if consulClient != nil {
		return nil
	} else {
		defaultConfig := api.DefaultConfig()

		defaultConfig.Address = fmt.Sprintf("%s:%d", app.GetApp().Consul.Address, app.GetApp().Consul.Port)
		defaultConfig.Scheme = app.GetApp().Scheme

		client, err := api.NewClient(defaultConfig)
		if err != nil {
			return err
		}

		consulClient = client

		return nil
	}
}

func findTrafficIp() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IP{}, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

// Finds any service that is registered for this address and port and returns it
func findServiceForHost(host string, port int) (*api.AgentService, error) {
	services, err := consulClient.Agent().ServicesWithFilter("Service == \"media-host-node\"")
	if err != nil {
		return nil, err
	}

	servicesForHost := make([]*api.AgentService, 0)

	for i := range services {
		service := services[i]

		h, ok := service.Meta["host"]
		if !ok {
			continue
		}

		p, ok := service.Meta["port"]
		if !ok {
			continue
		}

		metaPort, err := strconv.Atoi(p)
		if err != nil {
			continue
		}

		if h == host && metaPort == port {
			servicesForHost = append(servicesForHost, service)
		}
	}

	if len(servicesForHost) == 0 {
		return nil, nil
	}

	if len(servicesForHost) > 1 {
		return nil, fmt.Errorf("found multiple services for host %s and port %d", host, port)
	}

	return servicesForHost[0], nil
}

func RegisterService() error {
	trafficIp, err := findTrafficIp()
	if err != nil {
		return err
	}

	appInstance := app.GetApp()
	self := appInstance.SelfCfg

	service, err := findServiceForHost(trafficIp.String(), self.Port)
	if err != nil {
		return err
	}

	var nodeId uuid.UUID

	if service != nil {
		nodeId = uuid.MustParse(service.ID)
	} else {
		nodeId, err = uuid.NewUUID()
		if err != nil {
			return err
		}
	}

	registration := &api.AgentServiceRegistration{
		ID:      nodeId.String(),
		Name:    "media-host-node",
		Port:    self.Port,
		Address: trafficIp.String(),
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("%s://%s:%v/api/v1/health", self.Scheme, trafficIp, self.Port),
			Interval: "10s",
			Timeout:  "30s",
		},
		Meta: map[string]string{
			"scheme": self.Scheme,
			"host":   trafficIp.String(),
			"port":   strconv.Itoa(self.Port),
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return err
	}

	appInstance.NodeId = nodeId
	return nil
}

func UnregisterService() error {
	return consulClient.Agent().ServiceDeregister(app.GetApp().NodeId.String())
}
