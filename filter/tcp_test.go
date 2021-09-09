package filter

import (
	"log"
	"net"
	"testing"

	"github.com/sirupsen/logrus"
	tomb "gopkg.in/tomb.v1"
)

func TestFilter_ServeTCP(t *testing.T) {
	type fields struct {
		port string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"normal",
			fields{
				"65431",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := "This a chunk"
			chunk2 := "This a new chunk"
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				log.Fatal(err)
			}
			defer listener.Close()

			tombServer := tomb.Tomb{}

			go func() {
				defer tombServer.Done()
				conn, err := listener.Accept()
				if err != nil {
					select {
					case <-tombServer.Dying():
					default:
						t.Error("Failed to accept client")
						return
					}
					return
				}

				defer conn.Close()

				data := make([]byte, len(chunk))
				_, err = conn.Read(data)
				if err != nil {
					log.Fatal(err)
				}

				if string(data) != chunk {
					t.Errorf("got=%s, wants=%s", string(data), chunk)
				}

				_, err = conn.Write([]byte(chunk2))
				if err != nil {
					log.Fatal(err)
				}
			}()

			f := &Filter{
				url:  "tcp://" + listener.Addr().String(),
				port: tt.fields.port,
				log:  logrus.New(),
				kind: TCP,
			}

			doTest := make(chan interface{})
			tombClient := tomb.Tomb{}

			go func() {
				defer tombClient.Done()
				<-doTest
				conn2, err := net.Dial("tcp", "localhost:"+tt.fields.port)
				if err != nil {
					t.Errorf("error dialing remote addr: %v", err)
					return
				}

				defer conn2.Close()

				_, err = conn2.Write([]byte(chunk))
				if err != nil {
					log.Fatal(err)
				}

				<-doTest

				data := make([]byte, len(chunk2))
				_, err = conn2.Read(data)
				if err != nil {
					log.Fatal(err)
				}

				if string(data) != chunk2 {
					t.Errorf("got=%s, wants=%s", string(data), chunk2)
				}
			}()

			go f.ServeTCP()

			doTest <- ""
			tombServer.Wait()
			doTest <- ""
			tombClient.Wait()

		})
	}
}
