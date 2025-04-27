package health

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func healthz(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

// Start an health endpoint.
func Serve(upLog logrus.FieldLogger, port string) error {

	mux := http.NewServeMux()
	mux.HandleFunc("/", healthz)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		upLog.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}

	return err
}
