package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type config struct {
	grpcBind     string
	restBind     string
	smtpHost     string
	smtpPort     int
	smtpPassword string
	fromAddr     string
	notifiers    int
}

func (cfg *config) fromFlags() {
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

	if *grpcBind != "" {
		cfg.grpcBind = *grpcBind
	}
	if *restBind != "" {
		cfg.restBind = *restBind
	}
	if *smtpHost != "" {
		cfg.smtpHost = *smtpHost
	}
	if *smtpPort != 0 {
		cfg.smtpPort = *smtpPort
	}
	if *fromAddr != "" {
		cfg.fromAddr = *fromAddr
	}
	if *smtpPassword != "" {
		cfg.smtpPassword = *smtpPassword
	}
	if *notifiers != 0 {
		if *notifiers >= 100 {
			logs.warning.Println("Invalid value for flag 'notifiers', falling back to 5 (Allowed: 1-99)")
			cfg.notifiers = 5
		} else {
			cfg.notifiers = *notifiers
		}
	}
}
func (cfg *config) fromEnv() {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")

		if len(pair) > 1 {
			switch pair[0] {
			case "OTS_SMTP_HOST":
				cfg.smtpHost = pair[1]
				break
			case "OTS_SMTP_PORT":
				n, err := strconv.Atoi(pair[1])
				if err == nil && n > 0 && n < 65536 {
					cfg.smtpPort = n
				}
				break
			case "OTS_FROM_ADDR":
				cfg.fromAddr = pair[1]
				break
			case "OTS_SMTP_PASSWORD":
				cfg.smtpPassword = pair[1]
				break
			case "OTS_GRPC_BIND":
				cfg.grpcBind = pair[1]
				break
			case "OTS_REST_BIND":
				cfg.restBind = pair[1]
				break
			case "OTS_NOTIFIERS":
				n, err := strconv.Atoi(pair[1])
				if err == nil && n > 0 && n < 100 {
					cfg.notifiers = n
				} else {
					logs.warning.Println("Invalid value for environment variable 'OTS_NOTIFIERS' (Allowed: 1-99)")
				}
				break
			}
		}
	}
}
