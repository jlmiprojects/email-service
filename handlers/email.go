package handlers

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"blueassetgroup.com/email-service/repository"

	"blueassetgroup.com/email-service/models"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"gopkg.in/gomail.v2"
)

type Handler struct {
	nc        *nats.Conn
	config    *models.Config
	service   micro.Service
	logger    *slog.Logger
	dataStore *repository.MongoRepository
	temp      *template.Template
}

func NewHandler(version string, nc *nats.Conn, config *models.Config, r *repository.MongoRepository, temp *template.Template) (*Handler, error) {

	var err error

	handler := new(Handler)

	handler.nc = nc
	handler.config = config
	handler.dataStore = r
	handler.temp = temp

	// Setup ROOT logger for devices
	handler.logger = slog.With("name", "emailhandler")

	handler.service, err = micro.AddService(nc, micro.Config{
		Name:        "EmailService",
		Version:     version,
		Description: "Email Management Service",
	})

	if err != nil {
		slog.Error("Failed to add Service", "error", err)
	}

	err = handler.service.AddEndpoint("Send", micro.HandlerFunc(handler.Send), micro.WithEndpointSubject("send"))

	if err != nil {
		return nil, err
	}

	return handler, nil

}

func (handler Handler) Send(req micro.Request) {

	logger := handler.logger.With("name", "Send")

	err := func() error {

		logger.Info("New Request Recevied")

		request := new(models.Email)
		err := request.FromJSON(req.Data())
		if err != nil {
			return req.RespondJSON(models.Result{
				Message: "Failed to process payload :" + err.Error(),
				Error:   err.Error()})
		}

		if err := request.Validate(); err != nil {
			logger.Error("Validation Failed", "error", err)
			return req.RespondJSON(&models.Result{
				StatusCode: http.StatusBadRequest,
				Message:    "Validation Failed",
				Error:      err.Error(),
			})

		}

		var tpl bytes.Buffer

		if request.Data == nil {
			request.Data = make(map[string]any)
		}

		request.Template = strings.ToLower(request.Template)

		request.Template += ".html"

		err = handler.temp.ExecuteTemplate(&tpl, request.Template, request.Data)

		if err != nil {
			slog.Error("Failed to execute template, running default", "error", err)

			err = handler.temp.ExecuteTemplate(&tpl, "default.html", request.Data)

			if err != nil {
				slog.Error("Failed to execute template", "error", err)
				return req.RespondJSON(models.Result{
					Message: "Failed to execute template : " + err.Error(),
					Error:   err.Error()})

			}
		}

		email := models.Email{
			To:        request.To,
			From:      request.From,
			Subject:   request.Subject,
			Timestamp: time.Now(),
			Template:  tpl.String(), Data: request.Data,
		}

		err = handler.sendMail(request.From, request.To, request.Subject, tpl.String())

		if err != nil {
			return req.RespondJSON(models.Result{
				Message: "Failed to send email : " + err.Error(),
				Error:   err.Error()})
		}

		err = handler.dataStore.Save(&email)
		if err != nil {
			return req.RespondJSON(models.Result{
				Message: "Failed to add email to database : " + err.Error(),
				Error:   err.Error()})
		}

		return req.RespondJSON(models.MakeResult(200, "OK", nil))

	}()

	if err != nil {
		handler.logger.Error("Failed to reply", "error", err)
	}

}

func (handler Handler) sendMail(from string, to []string, subject string, message string) error {

	slog.Info("Sending Email", "host", handler.config.Email.Host,
		"port", handler.config.Email.Port,
		"from", from, "to", to, "subject", subject, "message", message)

	d := gomail.NewDialer(handler.config.Email.Host, handler.config.Email.Port,
		handler.config.Email.UserName, handler.config.Email.Password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)

	if err := d.DialAndSend(m); err != nil {
		slog.Error("Failed to send email", "error", err)
		return err
	}

	return nil

}
