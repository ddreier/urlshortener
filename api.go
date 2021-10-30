package main

import (
	"fmt"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
	bolt "go.etcd.io/bbolt"
	"io"
	"net/http"
	"net/url"
)

func router(db *bolt.DB) *bunrouter.Router {
	r := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware(reqlog.WithVerbose(true))))
	r.GET("/", func(w http.ResponseWriter, r bunrouter.Request) error {
		_, err := fmt.Fprintln(w, r.Method, r.Route(), r.Params().Map())
		return err
	})
	r.GET("/u/:id", getRedirect(db))
	r.PUT("/u/:id", putRedirect(db))
	r.POST("/u", postRedirect(db))

	return r
}

// Get redirected
func getRedirect(db *bolt.DB) func(w http.ResponseWriter, r bunrouter.Request) error {
	return func(w http.ResponseWriter, r bunrouter.Request) error {
		id := r.Param("id")

		var found bool
		var u string

		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("urls"))
			v := b.Get([]byte(id))

			if v != nil {
				found = true
				u = string(v)
			}

			return nil
		})
		if err != nil {
			return err
		}

		if found {
			http.Redirect(w, r.Request, u, http.StatusTemporaryRedirect)
			return nil
		}

		http.Error(w, fmt.Sprintf("URL for %q not found", id), http.StatusNotFound)
		return nil
	}
}

// Create/update a named redirect
func putRedirect(db *bolt.DB) func(w http.ResponseWriter, r bunrouter.Request) error {
	return func(w http.ResponseWriter, r bunrouter.Request) error {
		id := r.Param("id")

		lr := io.LimitReader(r.Body, 5000)
		body, err := io.ReadAll(lr)
		if err != nil {
			return err
		}

		u, err := url.Parse(string(body))
		if err != nil {
			http.Error(w, fmt.Sprintf("provided URL not valid: %s", err), http.StatusBadRequest)
			return nil
		}
		if u.Scheme == "" {
			http.Error(w, "provided URL not valid: scheme must be provided", http.StatusBadRequest)
			return nil
		}
		if u.Host == "" {
			http.Error(w, "provided URL not valid: host must be provided", http.StatusBadRequest)
			return nil
		}

		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("urls"))
			return b.Put([]byte(id), []byte(u.String()))
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error saving to database: %s", err), http.StatusBadRequest)
			return err
		}

		return nil
	}
}

// Add a new redirect with a random name
func postRedirect(db *bolt.DB) func(w http.ResponseWriter, r bunrouter.Request) error {
	return func(w http.ResponseWriter, r bunrouter.Request) error {
		lr := io.LimitReader(r.Body, 5000)
		body, err := io.ReadAll(lr)
		if err != nil {
			return err
		}

		u, err := url.Parse(string(body))
		if err != nil {
			http.Error(w, fmt.Sprintf("provided URL not valid: %s", err), http.StatusBadRequest)
			return nil
		}
		if u.Scheme == "" {
			http.Error(w, "provided URL not valid: scheme must be provided", http.StatusBadRequest)
			return nil
		}
		if u.Host == "" {
			http.Error(w, "provided URL not valid: host must be provided", http.StatusBadRequest)
			return nil
		}

		id := GenerateUnusedRedirectID(8, db)
		if id == "" {
			http.Error(w, "unable to generate random ID, please try again", http.StatusInternalServerError)
			return nil
		}

		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("urls"))
			return b.Put([]byte(id), []byte(u.String()))
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error saving to database: %s", err), http.StatusBadRequest)
			return err
		}

		_, _ = fmt.Fprintln(w, id)

		return nil
	}
}
