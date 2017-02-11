package imux

import (
	"net"
)

// A function that can be called by an IMUXSocket to reconnect after an error
type Redialer func() (net.Conn, error)

// A function that generates Redialers for specific bind addresses
type RedialerGenerator func(string) Redialer
