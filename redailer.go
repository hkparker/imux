package imux

import (
	"net"
)

type Redialer func() (net.Conn, error)
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
