package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var db *bolt.DB

type Email struct {
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	SignupDate time.Time `json:"signup"`
	Validated  bool      `json:"validated"`
}

type SimpleEmail struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func main() {

	var err error

	db, err = bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Emails"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	addr := ":8080"

	r := mux.NewRouter()

	r.Handle("/email", handlers.LoggingHandler(os.Stdout, allowedMethods(
		[]string{"OPTIONS", "GET", "POST", "PATCH", "DELETE"},
		handlers.MethodHandler{
			"GET":    http.HandlerFunc(GetEmails),
			"POST":   http.HandlerFunc(PostEmail),
			"PATCH":  http.HandlerFunc(PatchEmail),
			"DELETE": http.HandlerFunc(DeleteEmail),
		})))

	fmt.Printf("Starting server on port %v\n", addr)

	log.Fatal(http.ListenAndServe(":8080", r))

}

// GetEmails returns all emails currently held in the store in json format
func GetEmails(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Get Emails")

	return
}

// PostEmail submits a new email to the store
func PostEmail(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "PostEmail!")
	decoder := json.NewDecoder(r.Body)
	var e SimpleEmail
	err := decoder.Decode(&e.Email)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	fmt.Fprintln(w, e)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Emails"))
		id, _ := b.NextSequence()
		e.ID = int(id)
		err := b.Put([]byte("answer"), []byte("42"))
		return err
	})

	return
}

// PatchEmail alters an email, typically only used to validate submitted email address, but with changes to
// storing this might not be necessary if emails aren't stored until the email is replied to
func PatchEmail(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "PatchEmail")

	return
}

func DeleteEmail(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "DeleteEmail")

	return
}

func commaify(ss []string) (out string) {
	for i, s := range ss {
		out += s
		if i != len(ss)-1 {
			out += ","
		}
	}
	return
}

func allowedMethods(methods []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", commaify(methods))

		next.ServeHTTP(w, r)
	})
}
