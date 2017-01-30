package imux

import (
	"net"
)

// A function that can be called by an IMUXSocket to reconnect after an error
type Redialer func() (net.Conn, error)

// A function that generates Redialers for specific bind addresses
type RedialerGenerator func(string) Redialer

//func TLSRedailer() Redailer {
//	return func() (net.Conn, error) {
//		// Dial new conn
//		// Create a TLJ server for this new conn that
//		// accepts return chunks, validates session, writes
//		// chunk into write queue for correct socket
//		return nil, nil
//	}
//}
