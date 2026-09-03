package models

import "time"

type Logger struct {
	Level string `mapstructure:"level"`
	Type  string `mapstructure:"type"`
}

type Server struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type Config struct {
	TemplateDir string         `mapstructure:"templates"`
	StaticPath  string         `mapstructure:"static_path"`
	Logger      *Logger        `mapstructure:"logger"`
	Server      *Server        `mapstructure:"server"`
	Mongo       DatabaseConfig `mapstructure:"mongo"`
	Email       EmailConfig    `mapstructure:"email"`
	Nats        NatsConfig     `mapstructure:"nats"`
}

type DatabaseConfig struct {
	Uri             string        `mapstructure:"uri"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	EmailCollection string        `mapstructure:"email_collection"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
}

type EmailConfig struct {
	Port     int    `mapstructure:"port"`
	Host     string `mapstructure:"host"`
	UserName string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type NatsConfig struct {
	URI   string `mapstructure:"uri"`
	Token string `mapstructure:"token"`
}
