package grpcapi

import (
	"encoding/hex"
	"fmt"
	"net/mail"
	"net/url"
	"os"

	"gitlab.com/joukehofman/OTSthingy/proto"
	"gitlab.com/joukehofman/OTSthingy/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type TimestampServer struct {
	Requester *types.Requester
}

func (s TimestampServer) validateWithCallback(tsReq *OTSthingy.TimeStampRequest) error {
	// some validation
	if tsReq.EmailAddress == "" && tsReq.WebhookUrl == "" {
		return grpc.Errorf(codes.InvalidArgument, "One of EmailAddress or WebhookUrl is required")
	}
	if tsReq.EmailAddress != "" {
		_, err := mail.ParseAddress(tsReq.EmailAddress)
		if err != nil {
			return grpc.Errorf(codes.InvalidArgument, "Invalid email address")
		}
	}
	if tsReq.WebhookUrl != "" {
		_, err := url.ParseRequestURI(tsReq.WebhookUrl)
		if err != nil {
			return grpc.Errorf(codes.InvalidArgument, "Invalid webhook url")
		}
	}
	if len(tsReq.Label) > 100 {
		return grpc.Errorf(codes.InvalidArgument, "Label too long (max 100 characters)")
	}
	if len(tsReq.DocumentHash) != 32 {
		return grpc.Errorf(codes.InvalidArgument, "Please provide a 32-byte SHA256 hash")
	}
	return nil
}

func (s TimestampServer) WithCallback(ctx context.Context, tsReq *OTSthingy.TimeStampRequest) (*OTSthingy.IncompleteTimeStamp, error) {

	err := s.validateWithCallback(tsReq)

	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(fmt.Sprintf("%s.ots", hex.EncodeToString(tsReq.DocumentHash))); err == nil || !os.IsNotExist(err) {
		return nil, grpc.Errorf(codes.AlreadyExists, "Proof was already given")
	}

	result := OTSthingy.IncompleteTimeStamp{}
	s.Requester.AddRequest(tsReq)
	return &result, nil
}
