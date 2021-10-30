package main

import (
	"fmt"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
)

func main() {
	db, err := bolt.Open("urlshortener.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("urls"))

		err := tx.Bucket([]byte("urls")).Put([]byte("hacktoberfest"), []byte("https://hacktoberfest.digitalocean.com/"))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on :8888")
	log.Println(http.ListenAndServe(":8888", router(db)))
}

func router(db *bolt.DB) *bunrouter.Router {
	r := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware(reqlog.WithVerbose(true))))
	r.GET("/", func(w http.ResponseWriter, r bunrouter.Request) error {
		_, err := fmt.Fprintln(w, r.Method, r.Route(), r.Params().Map())
		return err
	})
	r.GET("/g/:id", func(w http.ResponseWriter, r bunrouter.Request) error {
		id := r.Param("id")

		var found bool
		var url string

		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("urls"))
			v := b.Get([]byte(id))

			if v != nil {
				found = true
				url = string(v)
			}

			return nil
		})
		if err != nil {
			return err
		}

		if found {
			http.Redirect(w, r.Request, url, http.StatusTemporaryRedirect)
			return nil
		}

		http.Error(w, fmt.Sprintf("URL for %q not found", id), http.StatusNotFound)
		return nil
	})

	return r
}
