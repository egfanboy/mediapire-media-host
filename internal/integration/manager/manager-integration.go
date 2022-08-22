package manager

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/egfanboy/mediapire-media-host/internal/app"

	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

type masterIntegration struct {
	app *app.App
}

func (i *masterIntegration) RegisterNode(masterScheme string, masterHost string, masterPort int) error {
	hostUri := fmt.Sprintf("%s:%v", masterHost, masterPort)

	net, err := findTrafficIp()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(registerRequest{Scheme: i.app.Scheme, Host: net.String(), Port: i.app.Port})

	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s://%s/api/v1/nodes/register", masterScheme, hostUri), "application/json", &buf)

	if err != nil {
		return fmt.Errorf("failed to register with media master due to: %s", err.Error())
	}

	// Return error if status code is over 3XX
	if resp != nil && resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusConflict {
			log.Info().Msg("Got a conflict when trying to register ourselves. Therefore, we are registered")

			return nil
		}
		return fmt.Errorf("media-host %s returned status code %d", hostUri, resp.StatusCode)
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

func NewManagerIntegration() ManagerApi {
	return &masterIntegration{app: app.GetApp()}
}
