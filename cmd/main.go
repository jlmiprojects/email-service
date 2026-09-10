package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"os"

	"blueassetgroup.com/email-service/handlers"
	"blueassetgroup.com/email-service/models"
	"blueassetgroup.com/email-service/repository"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

var (
	ServiceVersion string
)

func main() {

	ServiceVersion = "0.0.0"

	c := setupConfig()

	if c.Logger == nil {
		slog.Error("Logger is not set in config defaulting to DEBUG")
		c.Logger = &models.Logger{Level: "DEBUG"}
	}
	logger := newLogger(c.Logger)

	logger.Info("Logger setup info", "type", c.Logger)

	slog.Info("Logger is initialized")

	temp, err := template.ParseGlob(fmt.Sprintf("%s/*.html", c.TemplateDir))
	if err != nil {
		panic(err)
	}

	r, err := repository.NewMongo(c)

	if err != nil {
		panic(err)
	}

	nc, err := nats.Connect(c.Nats.URI, nats.Name("sms-client"), nats.Token(c.Nats.Token))

	if err != nil {
		panic(err)
	}

	logger.Info("NATS is connected on ", "uri", c.Nats.URI)

	_, err = handlers.NewHandler(ServiceVersion, nc, c, r, temp)

	if err != nil {
		panic(err)
	}

	slog.Info("Handlers configured and started")

	ctx := context.Background()

	<-ctx.Done()

}

func setupConfig() *models.Config {

	viper.SetConfigName("application") // Name of the file (without extension)
	viper.SetConfigType("json")        // Type of the configuration file
	viper.AddConfigPath("./conf")
	viper.AddConfigPath("../conf")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc") // Look for the file in the working directory

	// 2. Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// 3. Unmarshal the configuration into the struct
	var config models.Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	if config.Email.MaxAttachmentSize <= 0 {
		config.Email.MaxAttachmentSize = models.DefaultMaxAttachmentSize
	}
	if config.Email.MaxTotalAttachmentSize <= 0 {
		config.Email.MaxTotalAttachmentSize = models.DefaultMaxTotalAttachmentSize
	}

	return &config

}

func newLogger(logger *models.Logger) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(logger.Level)); err != nil {
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
