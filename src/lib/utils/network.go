package utils

import (
	"fmt"
	"net"

	"github.com/stormkit-io/stormkit-io/src/lib/errors"
)

// IsPortInUse checks if the given port is currently in use by any process.
func IsPortInUse(port int) bool {
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		// Port is not in use - wrap error for potential logging
		_ = errors.Wrap(err, errors.ErrorTypeExternal, fmt.Sprintf("port check failed for %s", addr))
		return false
	}

	if conn != nil {
		conn.Close()
	}

	return true
}
