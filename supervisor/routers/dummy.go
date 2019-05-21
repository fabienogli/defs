package routers

import (
	"github.com/gorilla/mux"
	"net/http"
	s "supervisor/storage"
	u "supervisor/utils"
)

func createFile(w http.ResponseWriter, r *http.Request) {

	client, err := s.NewLoadBalancerClient()
	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
	}
	defer client.Close()

	client.WhereTo("dummyhash", 120)

}

func queryFile(w http.ResponseWriter, r *http.Request) {

}

func SetDummyRoutes(r *mux.Router) *mux.Router {
	r.HandleFunc("/file", createFile).Methods("POST")
	r.HandleFunc("/file/{hash}/", queryFile).Methods("GET")
	return r
}
