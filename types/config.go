package types

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

var logs *Logger
var cfg *Config

func InitVars(l *Logger, c *Config) {
	logs = l
	cfg = c
}

// Config stores and reads settings
type Config struct {
	GrpcBind     string
	RestBind     string
	SmtpHost     string
	SmtpPort     int
	SmtpPassword string
	FromAddr     string
	Notifiers    int
}

func setIfNotEmpty(values map[*string]string) {
	for cfgVar, value := range values {
		if value != "" {
			*cfgVar = value
		}
	}
}

func (cfg *Config) FromFlags() {
	var (
		grpcBind     = flag.String("grpcBind", ":8181", "Expose storeapi GRPC on this port")
		restBind     = flag.String("restBind", ":8080", "Expose storeapi REST api on this port")
		smtpHost     = flag.String("smtpHost", "mail.entix.nl", "SMTP server host")
		smtpPort     = flag.Int("smtpPort", 587, "SMTP server port")
		smtpPassword = flag.String("smtpPassword", "", "SMTP password")
		fromAddr     = flag.String("fromAddr", "someone@example.com", "From address for notification emails")
		notifiers    = flag.Int("notifiers", 5, "Amount of notifier routines to run")
	)
	flag.Parse()
	setIfNotEmpty(map[*string]string{
		&cfg.GrpcBind:     *grpcBind,
		&cfg.RestBind:     *restBind,
		&cfg.SmtpHost:     *smtpHost,
		&cfg.FromAddr:     *fromAddr,
		&cfg.SmtpPassword: *smtpPassword,
	})
	if *smtpPort != 0 {
		cfg.SmtpPort = *smtpPort
	}
	if *notifiers != 0 {
		if *notifiers >= 100 {
			logs.Warning.Println("Invalid value for flag 'notifiers', falling back to 5 (Allowed: 1-99)")
			cfg.Notifiers = 5
		} else {
			cfg.Notifiers = *notifiers
		}
	}
}

func (cfg *Config) FromEnv() {
	cfgMap := map[string]*string{
		"SMTP_HOST":     &cfg.SmtpHost,
		"FROM_ADDR":     &cfg.FromAddr,
		"SMTP_PASSWORD": &cfg.SmtpPassword,
		"GRPC_BIND":     &cfg.GrpcBind,
		"REST_BIND":     &cfg.RestBind,
	}

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")

		// skip empty env vars or not starting with OTS_
		if len(pair) < 2 || len(pair[0]) < 5 || pair[0][:4] != "OTS_" {
			continue
		}
		if cfgVar, exists := cfgMap[pair[0][4:]]; exists {
			*cfgVar = pair[1]
		} else {
			switch pair[0] {
			case "OTS_SMTP_PORT":
				n, err := strconv.Atoi(pair[1])
				if err == nil && n > 0 && n < 65536 {
					cfg.SmtpPort = n
				}
				break
			case "OTS_NOTIFIERS":
				n, err := strconv.Atoi(pair[1])
				if err == nil && n > 0 && n < 100 {
					cfg.Notifiers = n
				} else {
					logs.Warning.Println("Invalid value for environment variable 'OTS_NOTIFIERS' (Allowed: 1-99)")
				}
				break
			}
		}
	}
}
