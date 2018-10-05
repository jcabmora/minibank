package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"minibank/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = getSecretKey()

// Registration username and password structure
type Registration struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserSessions holds a list of sessions to be associated with a user
type UserSessions struct {
	Sessions []string `json:"sessions"`
}

func (userSessions UserSessions) addItem(session uuid.UUID) []string {
	userSessions.Sessions = append(userSessions.Sessions, session.String())
	return userSessions.Sessions
}

// JWTToken token structure
type JWTToken struct {
	Token string `json:"token"`
}

// UserSessionCache  maps users to a list of sessions
var UserSessionCache = make(map[string]UserSessions)

// SessionUserCache maps sessions to a user
var SessionUserCache = make(map[string]string)

// ToJSON utility function to marshal Registation types
func (r Registration) ToJSON() string {
	jsonbytes, err := json.Marshal(r)
	if err != nil {
		log.Panic(err)
	}
	return string(jsonbytes)
}

// AuthValidationMiddleware validates token or cookies and forwards to next handler
func AuthValidationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		if authHeader != "" {
			authtoken := strings.Split(authHeader, " ")
			if len(authtoken) == 2 {
				token, err := jwt.Parse(authtoken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("invalid signing method")
					}
					return secretKey, nil
				})
				if err == nil {
					if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
						username := claims["username"].(string)
						exp := claims["exp"].(float64)
						// some kind of validation here is in order. For example, make sure user
						// has not been disabled
						if username != "" && int64(exp) > time.Now().Unix() {
							next(w, r)
							return
						}
					}
				}
			}
		} else {
			// attempt cookie validation
			cookie, err := r.Cookie("sessionid")
			if err == nil {
				session := cookie.Value
				if _, ok := SessionUserCache[session]; ok {
					// some kind of validation is also in order here.
					// We do not check for session expiration or session inactivity.
					next(w, r)
					return
				}

			}
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	})
}

// ServerUnavailableHandler handles requests when service is not available
func ServerUnavailableHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("Service Unavailable"))
}

// RegisterHandler handles registration requests
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
					lastID, _ := res.LastInsertId()
					w.Write([]byte(fmt.Sprintf("Successfully registered account %s", string(lastID))))
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

// LoginHandler handles cookie-based authentication
func LoginHandler(w http.ResponseWriter, r *http.Request) {
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
			w.Write([]byte("Unable to parse authentication data."))
		} else {
			var hashedPassword string
			row := models.Database.QueryRow("SELECT password FROM account WHERE username = ?", registration.Username)
			switch err := row.Scan(&hashedPassword); err {
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid Credentials"))
			case nil:
				// validate password by comparing
				err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(registration.Password))
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("Invalid Credentials"))
				} else {
					// create a new session and add cookie to response
					session, _ := uuid.NewRandom()
					http.SetCookie(w, &http.Cookie{
						Name:  "sessionid",
						Value: session.String(),
					})
					// Update the User to Session cache
					if userSessions, ok := UserSessionCache[registration.Username]; !ok {
						sessionList := []string{}
						userSessions = UserSessions{sessionList}
						userSessions.Sessions = userSessions.addItem(session)
						UserSessionCache[registration.Username] = userSessions
					} else {
						userSessions.Sessions = userSessions.addItem(session)
						UserSessionCache[registration.Username] = userSessions
					}
					// Update the Session to User cache
					SessionUserCache[session.String()] = registration.Username

				}
			default:
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Unable to validate credentials"))
			}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Method."))
	}
}

// TokenHandler handles token-based authentication requests
func TokenHandler(w http.ResponseWriter, r *http.Request) {
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
			w.Write([]byte("Unable to parse authentication data."))
		} else {
			var hashedPassword string
			row := models.Database.QueryRow("SELECT password FROM account WHERE username = ?", registration.Username)
			switch err := row.Scan(&hashedPassword); err {
			case sql.ErrNoRows:
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid Credentials"))
			case nil:
				// validate password by comparing
				err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(registration.Password))
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("Invalid Credentials"))
				} else {
					// generate a new token
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
						"username": registration.Username,
						"exp":      time.Now().Add(time.Hour * 24).Unix(),
					})
					tokenString, err := token.SignedString(secretKey)
					if err != nil {
						log.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("Unable to validate credentials"))
					} else {
						json.NewEncoder(w).Encode(JWTToken{Token: tokenString})
					}
				}
			default:
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Unable to validate credentials"))
			}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Method."))
	}
}

//SessionListHandler handles requests for listing user sessions. This function *should be chained by TokenValidatorMiddleware*
func SessionListHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("authorization")
	var userSessions UserSessions
	if authHeader != "" {
		authtoken := strings.Split(authHeader, " ")
		token, _ := jwt.Parse(authtoken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return secretKey, nil
		})
		claims := token.Claims.(jwt.MapClaims)
		username := claims["username"].(string)
		userSessions = UserSessionCache[username]
	} else {
		cookie, err := r.Cookie("sessionid")
		if err == nil {
			session := cookie.Value
			username := SessionUserCache[session]
			userSessions = UserSessionCache[username]
		}
	}
	json.NewEncoder(w).Encode(userSessions)

}

func getSecretKey() []byte {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if len(secretKey) == 0 {
		log.Panic("JWT_SECRET_KEY environment variable was not set")
	}
	return []byte(secretKey)
}
