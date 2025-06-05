package api

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type URL struct {
	ID        string    `bson:"_id" json:"id"`
	Original  string    `bson:"original_url" json:"original_url"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	ExpireAt  time.Time `bson:"expire_at" json:"expire_at"`
}

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
