package utils

// TODO irindul 2019-05-22 : This file is copy.paster from supervisor/utils, refactoring possible

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func Message(status bool, message string, code int) (map[string]interface{}) {
	return map[string]interface{}{"status": status, "message": message, "code": code}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(data["code"].(int))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	json.NewEncoder(w).Encode(data)
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func IsDevelopmentMode() bool {
	return os.Getenv("STORAGE_MODE") == "development"
}

func RespondWithError(w http.ResponseWriter, code int, err error) {
	var msg string
	if IsDevelopmentMode() {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}
	RespondWithJSON(w, code, map[string]string{"error": msg})

	go logError(code, err)
}

func RespondWithMsg(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"message": msg})
}

//// TODO irindul 2019-05-22 : Add other code we want to log, but its not necessary to log everything I guess
func logError(code int, err error) {
	if code == http.StatusInternalServerError {
		log.Printf("[%d] : %s", http.StatusInternalServerError, err.Error())
	}
}
