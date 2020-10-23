package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type statusHandler struct{}

func (s *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := json.Marshal(Status)
	if err != nil {
		w.Write([]byte("Internal Server Error"))
	}
	w.Write([]byte(result))
}

func MakeHTTPServer() *http.Server {
	portString := os.Getenv("PROBER_SERVER_PORT")
	port, err := strconv.ParseUint(portString, 10, 32)
	if err != nil {
		port = 9078
	} else {
		port = port
	}

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      &statusHandler{},
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return s
}
