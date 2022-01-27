package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/allegro/bigcache/v3"
	_ "github.com/mattn/go-sqlite3"
)

type Article struct {
	Title  string `json:"Title"`
	Author string `json:"author"`
	Link   string `json:"link"`
}

type JWToken struct {
	Token string `json:"Token"`
}

type LoginAttemp struct {
	Username string `json:"Username"`
	Code     string `json:"Code"`
	Secret   string `json:"Secret"`
}

type LoginPermit struct {
	Code string `json:"Code"`
	// UserId string `json:"UserId"`
}

type RegisterForm struct {
	Code     string `json:"Code"`
	Username string `json:"Code"`
	Secret   string `json:"Secret"`
}

var Articles []Article

type UserDB struct {
	db *sql.DB
	// driver string
	// file   string
}

func (userDB *UserDB) createTable() error {
	sql := `create table if not exists "users" (
	"id" integer primary key autoincrement,
	"username" text not null,
	"secret" text
	)`
	_, err := userDB.db.Exec(sql)
	return err
}

func (userDB *UserDB) putUser(u RegisterForm) bool {
	sql := `insert into users (username, secret) values(?,?)`
	stmt, err := userDB.db.Prepare(sql)
	if err != nil {
		return false
	}
	_, err = stmt.Exec(u.Username, u.Secret)
	if err != nil {
		return true
	} else {
		return false
	}
}

func (userDB *UserDB) checkUser(lg LoginAttemp) int {
	sql := `select first from users where username=?`
	stmt, err := userDB.db.Prepare(sql)
	if err != nil {
		return -1
	}
	res, err := stmt.Exec(lg.Username)
	if err != nil {
		return -1
	} else {
		rt, _ := res.LastInsertId()
		return int(rt)
	}

}

func homePage(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Articles)
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Articles)
}

func loginPage(cache *bigcache.BigCache, userDB *UserDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//post method
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			// errorResponse(w, "Content limit to application/json", http.StatusUnsupportedMediaType)
			return
		}
		var la LoginAttemp
		var unmarshalErr *json.UnmarshalTypeError
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&la)
		if err != nil {
			if errors.As(err, &unmarshalErr) {
				// errorResponse(w, "Bad request")
			} else {
				// errorResponse(w, "Bad request")
			}
			return
		}
		// valide
		// la.Username
		//
		token := JWToken{"------"}
		json.NewEncoder(w).Encode(token)
	}
}

func permitPage(cache *bigcache.BigCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// every get;
		// generate a byte[]
		byteArray := make([]byte, 20)
		rand.Read(byteArray)
		// conv string
		ss := base64.StdEncoding.EncodeToString(byteArray)
		// store in cache
		byteNull := make([]byte, 32)
		cache.Set(ss, byteNull)
		p := LoginPermit{ss}
		// p.Code = ss
		json.NewEncoder(w).Encode(p)
	}
}

func register(cache *bigcache.BigCache, d *UserDB) http.HandlerFunc {
	// initial db conn
	return func(w http.ResponseWriter, r *http.Request) {
		// load posted json of registerform
		// check db
		// register
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			// errorResponse(w, "Content limit to application/json", http.StatusUnsupportedMediaType)
			return
		}
		var rgu RegisterForm
		var unmarshalErr *json.UnmarshalTypeError
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&rgu)
		if err != nil {
			if errors.As(err, &unmarshalErr) {
				// errorResponse(w, "Bad request")
			} else {
				// errorResponse(w, "Bad request")
			}
			return
		}
		// valide
		la := LoginAttemp{
			Username: rgu.Username,
		}
		if d.checkUser(la) == -1 {
			return
		}
		// la.Username
		if d.putUser(rgu) {
			token := JWToken{"---------"}
			json.NewEncoder(w).Encode(token)
			return
		}
		//
	}
}

func handleRequest(cache *bigcache.BigCache, userDB *UserDB) {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/article", returnAllArticles)
	http.HandleFunc("/login", loginPage(cache, userDB))
	http.HandleFunc("/permit", permitPage(cache))
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func main() {
	Articles = []Article{
		Article{Title: "Python Intermediate and Advanced 101",
			Author: "Arkaprabha Majumdar",
			Link:   "https://www.amazon.com/dp/B089KVK23P"},
		Article{Title: "R programming Advanced",
			Author: "Arkaprabha Majumdar",
			Link:   "https://www.amazon.com/dp/B089WH12CR"},
		Article{Title: "R programming Fundamentals",
			Author: "Arkaprabha Majumdar",
			Link:   "https://www.amazon.com/dp/B089S58WWG"},
	}
	// instant db
	db, _ := sql.Open("sqlite3", "file.db3")
	userdb := &UserDB{db}
	userdb.createTable()
	// instant cache
	cache, _ := bigcache.NewBigCache(bigcache.DefaultConfig(2 * time.Minute))
	handleRequest(cache, userdb)
}
