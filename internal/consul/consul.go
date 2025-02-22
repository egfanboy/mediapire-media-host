package consul

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"

	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/hashicorp/consul/api"
)

const (
	mediahostConsulTag = "mediapire-media-host"
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

func generateNodeId(appName string) string {
	data := fmt.Sprintf("mediapire-media-host-%s", appName)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return hashStr[len(hashStr)-12:]
}

func RegisterService() error {
	appInstance := app.GetApp()

	selfIp := ""

	if appInstance.SelfCfg.Address != nil {
		selfIp = *appInstance.SelfCfg.Address
	} else {
		trafficIp, err := findTrafficIp()
		if err != nil {
			return err
		}
		selfIp = trafficIp.String()
	}

	self := appInstance.SelfCfg

	nodeId := generateNodeId(appInstance.Name)

	registration := &api.AgentServiceRegistration{
		ID:      nodeId,
		Name:    appInstance.Name,
		Port:    self.Port,
		Address: selfIp,
		Tags:    []string{mediahostConsulTag},
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("%s://%s:%v/api/v1/health", self.Scheme, selfIp, self.Port),
			Interval: "10s",
			Timeout:  "30s",
		},
		Meta: map[string]string{
			"scheme": self.Scheme,
			"host":   selfIp,
			"port":   strconv.Itoa(self.Port),
		},
	}

	err := consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return err
	}

	appInstance.NodeId = nodeId
	return nil
}

func UnregisterService() error {
	return consulClient.Agent().ServiceDeregister(app.GetApp().NodeId)
}
