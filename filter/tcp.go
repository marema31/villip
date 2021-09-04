package filter

import (
	"io"
	"net"
)

func (f *Filter) ServeTCP() error {
	localAddr := ":" + f.port
	remoteAddr := f.url[len("tcp://"):]

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		log := f.log.WithField("remote", conn.RemoteAddr())

		log.Debug("New connection")

		if err != nil {
			log.Errorf("error accepting connection: %v", err)

			continue
		}

		go func() {
			defer conn.Close()

			conn2, err := net.Dial("tcp", remoteAddr)
			if err != nil {
				log.Errorf("error dialing remote addr: %v", err)

				return
			}

			defer conn2.Close()

			closer := make(chan struct{}, 2) //nolint: gomnd

			go copyTCP(closer, conn2, conn)
			go copyTCP(closer, conn, conn2)
			<-closer
			log.Debug("Connection complete")
		}()
	}
}

func copyTCP(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}
