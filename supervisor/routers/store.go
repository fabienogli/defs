package routers

import (
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	//s "supervisor/storage"
	//u "supervisor/utils"
)

func upload(w http.ResponseWriter, r *http.Request) {

}

func download(w http.ResponseWriter, r *http.Request) {
	// TODO irindul 2019-05-26 : Get storage from loadbalancer
	storateDns := "http://storage"
	storagePortStr := os.Getenv("STORAGE_PORT")

	vars := mux.Vars(r)
	hash := vars["hash"]

	url := storateDns + ":" + storagePortStr + "/download/" + hash

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
