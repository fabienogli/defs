package routers

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	u "supervisor/utils"
	"supervisor/models"
	//jwt "github.com/dgrijalva/jwt-go"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}


func login(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		u.RespondWithError(w, http.StatusUnprocessableEntity, err)
	}

	//Check database for username/pseudo 
	//if one is not found ==> 404 + Pair username/pseudo is wrong
	//c.Sign()

}

func signin(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		u.RespondWithError(w, http.StatusUnprocessableEntity, err)
	}

	c, err := models.NewClient(creds.Username, creds.Password);
	if err != nil {
		//@todo respond with proper error
	}

	json.NewEncoder(w).Encode(c)
}

func SetAuthenticationRoutes (r *mux.Router) *mux.Router {
	r.HandleFunc("/token", login).Methods("POST")
	r.HandleFunc("/users", signin).Methods("POST")
	return r
}