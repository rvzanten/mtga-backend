package types

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"

	"github.com/scorredoira/email"
)

// Notifier sends notifications for completed timestamping requests
type Notifier struct {
}

func (notify *Notifier) callWebhook(req *Request, txid []byte) error {
	return httpPost(req.tsRequest.WebhookUrl, map[string]string{
		"label":  req.tsRequest.Label,
		"txhash": hex.EncodeToString(txid),
		"proof":  hex.EncodeToString(req.proof),
	})
}

func (notify *Notifier) sendMail(req *Request, txid []byte) error {
	logs.Debug.Println("Sending mail to " + req.tsRequest.EmailAddress)
	content := fmt.Sprintf(
		"We would like to notify that your timestamp request (labelled '%s') has been finalized. The transaction hash:\r\n"+
			"\r\n"+
			"%s",
		req.tsRequest.Label,
		hex.EncodeToString(txid),
	)
	addr, err := mail.ParseAddress(req.tsRequest.EmailAddress)
	if err != nil {
		return err
	}
	msg := email.NewMessage("timestamp request adopted into Bitcoin blockchain", content)
	msg.From = *addr
	msg.To = []string{req.tsRequest.EmailAddress}
	err = msg.AttachBuffer(hex.EncodeToString(req.tsRequest.DocumentHash)+".proof", req.proof, false)
	if err != nil {
		return err
	}
	auth := smtp.PlainAuth("", cfg.FromAddr, cfg.SmtpPassword, cfg.SmtpHost)
	err = email.Send(fmt.Sprintf("%s:%d", cfg.SmtpHost, cfg.SmtpPort), auth, msg)

	return err
}

func httpPost(URL string, params map[string]string) (err error) {
	//build url
	u, err := url.Parse(URL)

	if err != nil {
		return err
	}
	requestURL := fmt.Sprintf("%v", u)
	//prepare params
	data := url.Values{}
	//convert params into querystring
	if len(params) > 0 {
		for k, v := range params {
			data.Add(k, v)
		}
	}
	//create request
	_, err = http.PostForm(requestURL, data)
	if err != nil {
		return err
	}
	return nil
}
