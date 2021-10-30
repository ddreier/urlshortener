package main

import (
	bolt "go.etcd.io/bbolt"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

// GenerateUnusedRedirectID will try 10 times to generate an ID that doesn't exist in the DB
func GenerateUnusedRedirectID(n int, db *bolt.DB) string {
	var id string
	tries := 10
	unused := false
	for !unused && tries > 0 {
		id = RandStringBytesRmndr(n)

		_ = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("urls"))
			if v := b.Get([]byte(id)); v == nil {
				unused = true
			}
			return nil
		})

		tries--
	}

	return id
}
