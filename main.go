package main

import (
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
	"time"
)

func main() {
	db, err := bolt.Open("urlshortener.db", 0666, &bolt.Options{Timeout: 5 * time.Second})
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
