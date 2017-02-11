package main

import (
	"net/mail"
	"net/url"

	"gitlab.com/joukehofman/OTSthingy/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type timestampServer struct {
	requestr *requester
}

func (s timestampServer) WithCallback(ctx context.Context, tsReq *OTSthingy.TimeStampRequest) (*OTSthingy.IncompleteTimeStamp, error) {

	// some validation
	if tsReq.EmailAddress == "" && tsReq.WebhookUrl == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "One of EmailAddress or WebhookUrl is required")
	}
	if tsReq.EmailAddress != "" {
		_, err := mail.ParseAddress("Alice <alice@example.com>")
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "Invalid email address")
		}
	}
	if tsReq.WebhookUrl != "" {
		_, err := url.ParseRequestURI(tsReq.WebhookUrl)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "Invalid webhook url")
		}
	}

	result := OTSthingy.IncompleteTimeStamp{}
	s.requestr.addRequest(tsReq)
	return &result, nil
}
