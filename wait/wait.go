package wait

import (
	"net"
	"time"

	"github.com/cockroachdb/errors"
)

// Wait tests availability of a service
// via pinging the given tcp address.
func Wait(tcpAddr string, timeout time.Duration) error {
	then := time.Now()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		conn, err := net.DialTimeout("tcp", tcpAddr, 3*time.Second)
		if err != nil {
			if time.Now().After(then.Add(timeout)) {
				return errors.Wrapf(err, "couldnt connect to %s", tcpAddr)
			}

			continue
		}

		err = conn.Close()
		if err != nil {
			return errors.Wrap(
				err,
				"couldnt close connection")
		}

		return nil
	}

	return nil
}
