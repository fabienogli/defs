package routers

import (
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	s "supervisor/storage"
	u "supervisor/utils"
)

func upload(w http.ResponseWriter, r *http.Request) {

}


func download(w http.ResponseWriter, r *http.Request) {
	// TODO irindul 2019-05-26 : Get storage from loadbalancer


	vars := mux.Vars(r)
	hash := vars["hash"]

	lb, err := s.NewLoadBalancerClient()
	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	response := lb.WhereIs(hash)

	// TODO irindul 2019-05-28 : Parse properly
	storeDns := strings.Split(response, " ")[1]
	storagePortStr := os.Getenv("STORAGE_PORT")
	// TODO irindul 2019-05-28 : Get http protocol frop .env (so we can easilly switch to https)
	url := "http://" + storeDns + ":" + storagePortStr + "/download/" + hash

	proxyRequest, err := http.NewRequest(http.MethodGet, url, r.Body)

	// TODO irindul 2019-05-26 : Change this with env variable
	proxyRequest.Header.Set("HOST", "supervisor")


	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyRequest.Header = make(http.Header)
	for h, val := range r.Header {
		proxyRequest.Header[h] = val
	}

	client := &http.Client{}
	resp, err := client.Do(proxyRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	//u.RespondWithMsg(w, http.StatusOK, "swag")
}

func SetStoreRoute(r *mux.Router) *mux.Router {
	r.HandleFunc("/file", createFile).Methods("POST")
	r.HandleFunc("/file/{hash}", download).Methods("GET")
	return r
}
