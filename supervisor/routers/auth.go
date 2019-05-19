package routers

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Auth working")
}

func SetAuthenticationRoutes (r *mux.Router) *mux.Router {
	r.HandleFunc("/auth", login)
	return r
}