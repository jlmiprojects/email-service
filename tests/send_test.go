package tests

// Integration tests that exercise the email service over NATS.
//
// They require a running NATS server (see conf/application.json -> nats.uri)
// and a running instance of the email service (go run cmd/main.go).
//
// Run just these tests with:
//
//	go test ./tests/ -run TestSend -v
//
// They are skipped automatically under `go test -short` or when NATS is
// unreachable, so they don't break `go test ./...`.

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"blueassetgroup.com/email-service/models"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

// sendSubject is the subject the service's micro endpoint listens on
// (handlers.NewHandler -> AddEndpoint("Send", ..., WithEndpointSubject("send"))).
const sendSubject = "send"

// loadNatsConfig reads nats.uri / nats.token from conf/application.json,
// the same file cmd/main.go loads.
func loadNatsConfig(t *testing.T) models.NatsConfig {
	t.Helper()

	v := viper.New()
	v.SetConfigName("application")
	v.SetConfigType("json")
	v.AddConfigPath("../conf")
	v.AddConfigPath("./conf")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("reading config: %v", err)
	}

	var c models.Config
	if err := v.Unmarshal(&c); err != nil {
		t.Fatalf("decoding config: %v", err)
	}
	return c.Nats
}

// connect dials NATS, skipping the test (rather than failing) when no
// server is available - these are opt-in integration tests.
func connect(t *testing.T) *nats.Conn {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping NATS integration test in -short mode")
	}

	cfg := loadNatsConfig(t)

	nc, err := nats.Connect(cfg.URI,
		nats.Name("email-service-test"),
		nats.Token(cfg.Token),
		nats.Timeout(2*time.Second),
	)
	if err != nil {
		t.Skipf("NATS not reachable at %s: %v", cfg.URI, err)
	}

	t.Cleanup(nc.Close)
	return nc
}

// send publishes an Email request on the `send` subject and returns the
// service's decoded Result.
func send(t *testing.T, nc *nats.Conn, email models.Email) models.Result {
	t.Helper()

	payload, err := json.Marshal(email)
	if err != nil {
		t.Fatalf("marshalling email: %v", err)
	}

	t.Logf("-> %s %s", sendSubject, payload)

	msg, err := nc.Request(sendSubject, payload, 15*time.Second)
	if errors.Is(err, nats.ErrNoResponders) {
		t.Skipf("no responder on %q - start the service with `go run cmd/main.go`", sendSubject)
	}
	if err != nil {
		t.Fatalf("NATS request failed: %v", err)
	}

	t.Logf("<- %s", msg.Data)

	var result models.Result
	if err := result.FromJSON(msg.Data); err != nil {
		t.Fatalf("decoding result %q: %v", msg.Data, err)
	}
	return result
}

// testAddr returns the address to send test mail to/from. Override with
// EMAIL_TEST_TO / EMAIL_TEST_FROM to avoid actually mailing the default box.
func testAddr(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return fallback
}

func TestSendEmail_DefaultTemplate(t *testing.T) {
	nc := connect(t)

	from := testAddr("EMAIL_TEST_FROM", "jacques@blueassetgroup.com")
	to := testAddr("EMAIL_TEST_TO", "jacques@blueassetgroup.com")

	result := send(t, nc, models.Email{
		From:    from,
		To:      []string{to},
		Subject: "Broker portal email service - default template test",
		// No Template -> service falls back to default.html.
	})

	if result.Error != "" {
		t.Fatalf("service returned error: %s (%s)", result.Error, result.Message)
	}
	if result.StatusCode != 0 && result.StatusCode != 200 {
		t.Fatalf("unexpected status code %d: %s", result.StatusCode, result.Message)
	}
}

func TestSendEmail_WithAttachment(t *testing.T) {
	nc := connect(t)

	from := testAddr("EMAIL_TEST_FROM", "jacques@blueassetgroup.com")
	to := testAddr("EMAIL_TEST_TO", "jacques@blueassetgroup.com")

	result := send(t, nc, models.Email{
		From:    from,
		To:      []string{to},
		Subject: "Broker portal email service - attachment test",
		Attachments: []models.Attachment{
			{
				Filename:    "note.txt",
				ContentType: "text/plain",
				Content:     []byte("hello from the email-service attachment test\n"),
			},
		},
	})

	if result.Error != "" {
		t.Fatalf("service returned error: %s (%s)", result.Error, result.Message)
	}
	if result.StatusCode != 0 && result.StatusCode != 200 {
		t.Fatalf("unexpected status code %d: %s", result.StatusCode, result.Message)
	}
}

// The templates the broker portal (ui/services/email.go) sends by name.
func TestSendEmail_BrokerPortalTemplates(t *testing.T) {
	nc := connect(t)

	from := testAddr("EMAIL_TEST_FROM", "jacques@blueassetgroup.com")
	to := testAddr("EMAIL_TEST_TO", "jacques@blueassetgroup.com")

	cases := []struct {
		name    string
		subject string
		data    map[string]any
	}{
		{
			name:    "profile_invite",
			subject: "Broker portal - profile_invite template test",
			data: map[string]any{
				"username":      "jbloggs",
				"temp_password": "Xy7-Kp2Qa9",
				"login_url":     "http://localhost:8888/",
			},
		},
		{
			name:    "password_reset",
			subject: "Broker portal - password_reset template test",
			data: map[string]any{
				"username":      "jbloggs",
				"temp_password": "Zq4-Lm8Rb1",
				"login_url":     "http://localhost:8888/",
			},
		},
		{
			name:    "quote_accept_invite",
			subject: "Broker portal - quote_accept_invite template test",
			data: map[string]any{
				"client_name":  "Jane Client",
				"advisor_name": "Alex Advisor",
				"plan_name":    "Retirement Annuity",
				"amount":       "R 1,500.00",
				"accept_url":   "http://localhost:8888/q/test-token",
				"expires_at":   "17 Sep 2026",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := send(t, nc, models.Email{
				From:     from,
				To:       []string{to},
				Subject:  tc.subject,
				Template: tc.name,
				Data:     tc.data,
			})

			if result.Error != "" {
				t.Fatalf("service returned error: %s (%s)", result.Error, result.Message)
			}
			if result.StatusCode != 0 && result.StatusCode != 200 {
				t.Fatalf("unexpected status code %d: %s", result.StatusCode, result.Message)
			}
		})
	}
}
