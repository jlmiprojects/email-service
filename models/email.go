package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	Content     []byte `json:"content,omitempty"` // encoding/json base64-encodes []byte automatically
	Size        int64  `json:"size,omitempty"`    // populated on the persisted (metadata-only) copy
}

type Email struct {
	ID          primitive.ObjectID `bson:"_id"`
	Timestamp   time.Time          `json:"timestamp"`
	From        string             `json:"from"`
	To          []string           `json:"to"`
	Subject     string             `json:"subject"`
	Data        any                `json:"data"`
	Template    string             `json:"template"`
	Attachments []Attachment       `json:"attachments,omitempty"`
}

func (o *Email) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

func (u Email) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u Email) Validate() error {
	for i, a := range u.Attachments {
		if a.Filename == "" {
			return fmt.Errorf("attachment %d: filename is required", i)
		}
		if len(a.Content) == 0 {
			return fmt.Errorf("attachment %q: content is empty", a.Filename)
		}
		if strings.ContainsAny(a.Filename, `/\`) || strings.Contains(a.Filename, "..") {
			return fmt.Errorf("attachment %q: filename must not contain path separators", a.Filename)
		}
	}
	return nil
}
