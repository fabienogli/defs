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

func checkForUnprocessable(err error, w http.ResponseWriter) {
	if err != nil {
		code := http.StatusUnprocessableEntity
		var msg string
		if u.IsDevelopmentMode() {
			msg = err.Error()
		} else {
			msg = http.StatusText(code)
		}

		u.RespondWithError(w, code, msg)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	checkForUnprocessable(err, w)

	//Check database for username/pseudo 
	//if one is not found ==> 404 + Pair username/pseudo is wrong
	//c.Sign()

}

func signin(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	checkForUnprocessable(err, w)

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