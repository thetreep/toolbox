package transport

import (
	"net"
	"time"

	"github.com/cockroachdb/errors"
	externalip "github.com/glendc/go-external-ip"
)

// GetPublicIP returns when possible the current public ip.
func GetPublicIP() (net.IP, error) {
	ip, err := externalip.DefaultConsensus(
		externalip.DefaultConsensusConfig().WithTimeout(1*time.Second), nil).ExternalIP()
	if err != nil {
		return nil, errors.WithStack(errors.Wrap(err, "could not get public ip"))
	}

	return ip, nil
}

// GetLocalIP returns the non loopback local IP of the host.
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.WithStack(errors.Wrap(err, "could not get local ip"))
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("couldnt not find valid interface to select local ip")
}
