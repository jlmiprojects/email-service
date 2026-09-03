package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Email struct {
	ID        primitive.ObjectID `bson:"_id"`
	Timestamp time.Time          `json:"timestamp"`
	From      string             `json:"from"`
	To        []string           `json:"to"`
	Subject   string             `json:"subject"`
	Data      any                `json:"data"`
	Template  string             `json:"template"`
}

func (o *Email) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

func (u Email) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u Email) Validate() error {
	return nil
}
