package routers

import (
	"github.com/gorilla/mux"
	"net/http"
	s "supervisor/storage"
)

/*
var ip = os.Getenv("LOADBALANCER_IP")
var port = os.Getenv("LOADBALANCER_PORT")
*/

func createFile(w http.ResponseWriter, r *http.Request) {

	client, _ := s.NewLoadBalancerClient("127.0.0.1", 10001)
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
