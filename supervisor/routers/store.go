package routers

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	s "supervisor/storage"
	u "supervisor/utils"
)

func upload(w http.ResponseWriter, r *http.Request) {

}

func download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	lb, err := s.NewLoadBalancerClient()
	defer lb.Close()
	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	response, err := lb.WhereIs(hash)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			u.RespondWithError(w, http.StatusGatewayTimeout, err)
			return
		}
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respPart := strings.Split(response, " ")

	if respPart[0] != s.Ok.String() {
		switch respPart[0] {
		case s.HashNotFound.String():
			u.RespondWithError(w, http.StatusNotFound, fmt.Errorf("hash not found"))
			return
		default:
			u.RespondWithError(w, http.StatusNotImplemented, fmt.Errorf("response not implemented : %s", respPart[0]))
		}
	}

	storeDns := respPart[1]
	storagePortStr := os.Getenv("STORAGE_PORT")
	protocol := os.Getenv("STORAGE_PROTOCOL")
	url := protocol + "://" + storeDns + ":" + storagePortStr + "/download/" + hash

	proxyRequest, err := http.NewRequest(http.MethodGet, url, r.Body)
	hostName := os.Getenv("SUPERVISOR_HTTP_HOST")
	proxyRequest.Header.Set("HOST", hostName)


	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyRequest.Header = make(http.Header)
	for h, val := range r.Header {
		proxyRequest.Header[h] = val
	}

	client := &http.Client{}
	resp, err := client.Do(proxyRequest)
	if err != nil {
		u.RespondWithError(w, http.StatusBadGateway, err)
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func
SetStoreRoute(r *mux.Router) *mux.Router {
	r.HandleFunc("/file", createFile).Methods("POST")
	r.HandleFunc("/file/{hash}", download).Methods("GET")
	return r
}
