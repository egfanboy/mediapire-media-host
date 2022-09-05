package manager

import (
	"context"
	"fmt"

	"net"
	"net/http"

	api "github.com/egfanboy/mediapire-manager/pkg/api"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/egfanboy/mediapire-media-host/internal/app"

	"github.com/rs/zerolog/log"
)

type masterIntegration struct {
	app           *app.App
	ManagerClient api.MediaManagerApi
}

func (i *masterIntegration) RegisterNode() error {
	net, err := findTrafficIp()
	if err != nil {
		return err
	}

	selfCfg := i.app.SelfCfg

	resp, err := i.ManagerClient.RegisterNode(types.RegisterNodeRequest{Host: net, Scheme: selfCfg.Scheme, Port: &selfCfg.Port})

	if err != nil {
		return err
	}

	// Return error if status code is over 3XX
	if resp != nil && resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusConflict {
			log.Info().Msg("Got a conflict when trying to register ourselves. Therefore, we are registered")

			return nil
		}
		return fmt.Errorf("media-host %s returned status code %d", i.app.Manager.Host, resp.StatusCode)
	}

	return nil
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

func NewManagerIntegration(ctx context.Context) ManagerApi {
	app := app.GetApp()

	manager := app.Manager

	host := manager.Host

	if host == "localhost" {
		host = "127.0.0.1"
	}

	return &masterIntegration{app: app, ManagerClient: api.NewManagerClient(ctx, api.ManagerConnectionConfig{
		Host:   net.ParseIP(host),
		Scheme: manager.Scheme,
		Port:   &manager.Port,
	})}
}
