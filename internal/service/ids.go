package service

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func newID() string {
	id, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
	if err != nil {
		// Extremely unlikely with crypto/rand.
		panic(err)
	}
	return id.String()
}
