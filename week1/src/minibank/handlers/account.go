package handlers

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"minibank/models"
	"net/http"
	"time"
)

type Registration struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r Registration) ToJSON() string {
	jsonbytes, err := json.Marshal(r)
	if err != nil {
		log.Panic(err)
	}
	return string(jsonbytes)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Unable to read request body."))
		}
		registration := Registration{}
		err = json.Unmarshal(body, &registration)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Unable to parse registration data."))
		} else {
			strlength := len(registration.Password)
			if strlength >= 10 {
				hashedpw, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Unable to process request."))
				}
				res, err := models.Database.Exec("INSERT INTO account(username, password, timestamp) VALUES (?, ?, ?)",
					registration.Username,
					hashedpw,
					time.Now().UnixNano()/1000000)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Unable to register new account"))
				} else {
					last_id, _ := res.LastInsertId()
					w.Write([]byte(fmt.Sprintf("Successfully registered account %s", string(last_id))))
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Passwords must be at least 10 characters long"))
			}
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Method."))
	}
}
