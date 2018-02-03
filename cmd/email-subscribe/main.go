package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/badoux/checkmail"
	"github.com/boltdb/bolt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/natethinks/email-subscribe/internal/respond"
	"github.com/natethinks/email-subscribe/internal/utility"
)

var (
	db *bolt.DB
)

type Email struct {
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	SignupDate time.Time `json:"signup"`
	Validated  bool      `json:"validated"`
}

type SimpleEmail struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func main() {

	// flags
	delete := flag.Bool("D", false, "clear all entries from db")
	//clean := flag.Bool("C", false, "clean all partial entries from the db")
	flag.Parse()

	var err error

	db, err = bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *delete {
		if c := utility.AskForConfirmation("Do you really want to delete the entire database?"); c {
			fmt.Println("You absolute mad man!")
			fmt.Println("Store has been deleted and will be recreated")
			db.Update(func(tx *bolt.Tx) error {
				return tx.DeleteBucket([]byte("Emails"))
			})
		}
	}

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

	var buffer bytes.Buffer

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("Emails")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			s := fmt.Sprintf("key=%s, value=%s\n", k, v)
			buffer.WriteString(s)
		}
		return nil
	})

	respond.JSON(w, buffer.String())
	defer r.Body.Close()
	return
}

// PostEmail submits a new email to the store
func PostEmail(w http.ResponseWriter, r *http.Request) {

	var e SimpleEmail
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&e)
	if err != nil {
		panic(err)
		w.WriteHeader(http.StatusBadRequest)
		respond.JSON(w, err)
		return
	}
	defer r.Body.Close()

	err = checkmail.ValidateFormat(e.Email)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		respond.JSON(w, err)
		return
	}

	err = checkmail.ValidateHost(e.Email)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		respond.JSON(w, err)
		return
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Emails"))
		id, _ := b.NextSequence()
		e.ID = strconv.Itoa(int(id))
		err := b.Put([]byte(e.ID), []byte(e.Email))
		return err
	})

	w.WriteHeader(http.StatusOK)
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

func List(bucket string) {
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
}
