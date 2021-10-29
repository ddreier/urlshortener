package main

import (
	"fmt"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"log"
	"net/http"
)

func main() {
	router := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware(reqlog.WithVerbose(true))))
	router.GET("/", func(w http.ResponseWriter, r bunrouter.Request) error {
		_, err := fmt.Fprintln(w, r.Method, r.Route(), r.Params().Map())
		return err
	})
	router.GET("/g/:id", func(w http.ResponseWriter, r bunrouter.Request) error {
		if r.Param("id") == "hacktoberfest" {
			http.Redirect(w, r.Request, "https://hacktoberfest.digitalocean.com/", http.StatusTemporaryRedirect)
			return nil
		}

		http.Error(w, fmt.Sprintf("URL for %q not found", r.Param("id")), http.StatusNotFound)
		return nil
	})

	log.Println("Listening on :8888")
	log.Println(http.ListenAndServe(":8888", router))
}
