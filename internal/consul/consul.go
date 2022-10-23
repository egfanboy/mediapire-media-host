package consul

import (
	"fmt"
	"net"

	"github.com/egfanboy/mediapire-media-host/internal/app"
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

func RegisterService() error {
	trafficIp, err := findTrafficIp()

	if err != nil {
		return err
	}

	self := app.GetApp().SelfCfg

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("media-host-node-%s", trafficIp),
		Name:    "media-host-node",
		Port:    self.Port,
		Address: trafficIp.String(),
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("%s://%s:%v/api/v1/health", self.Scheme, trafficIp, self.Port),
			Interval: "10s",
			Timeout:  "30s",
		},
	}

	return consulClient.Agent().ServiceRegister(registration)
}

func UnregisterService() error {

	trafficIp, err := findTrafficIp()

	if err != nil {
		return err
	}

	return consulClient.Agent().ServiceDeregister(fmt.Sprintf("media-host-node-%s", trafficIp))
}
