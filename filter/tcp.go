package filter

import (
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

// ServeTCP listen on the given port and start a goroutine to handle each connection.
func (f *Filter) ServeTCP() error {
	localAddr := ":" + f.port
	remoteAddr := f.url[len("tcp://"):]

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return err
	}

	for {
		clientConn, err := listener.Accept()
		log := f.log.WithField("remote", clientConn.RemoteAddr())

		log.Debug("New connection")

		if err != nil {
			log.Errorf("error accepting connection: %v", err)

			continue
		}

		go func() {
			defer clientConn.Close()

			serverConn, err := net.Dial("tcp", remoteAddr)
			if err != nil {
				log.Errorf("error dialing remote addr: %v", err)

				return
			}

			defer serverConn.Close()

			closer := make(chan struct{}, 2)

			go copyTCP(closer, clientConn, serverConn, log.WithField("type", "response"))
			go copyTCP(closer, serverConn, clientConn, log.WithField("type", "request"))
			<-closer
			log.Debug("Connection complete")
		}()
	}
}

func copyTCP(closer chan struct{}, dst io.Writer, src io.Reader, log logrus.FieldLogger) {
	n, err := io.Copy(dst, src)
	if err != nil {
		log.Errorf("transfer fail: %v", err)
	}

	log.Debugf("transferred: %d bytes", n)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
