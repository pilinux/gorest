package service

import (
	"context"

	"github.com/mrz1836/postmark"
)

// PostmarkParams - parameters for postmark email delivery service
type PostmarkParams struct {
	ServerToken   string
	TemplateID    int64
	From          string
	To            string
	Tag           string
	TrackOpens    bool
	TrackLinks    string
	MessageStream string
	HTMLModel     map[string]interface{}
}

// Postmark email delivery service using HTML templates
//
// https://postmarkapp.com/developer/api/email-api
//
// https://postmarkapp.com/developer/api/templates-api
//
// https://account.postmarkapp.com/servers/{ServerID}/templates
func Postmark(params PostmarkParams) (postmark.EmailResponse, error) {
	client := postmark.NewClient(params.ServerToken, "")

	email := postmark.TemplatedEmail{}
	email.TemplateID = params.TemplateID
	email.From = params.From
	email.To = params.To
	email.Tag = params.Tag
	email.TrackOpens = params.TrackOpens
	email.TrackLinks = params.TrackLinks
	email.MessageStream = params.MessageStream
	email.TemplateModel = params.HTMLModel

	res, err := client.SendTemplatedEmail(context.Background(), email)
	/*
		res.To:				recipient email address
		res.SubmittedAt:	timestamp in UTC
		res.MessageID:		ID of message
		res.ErrorCode:		postmark API Error Codes [https://postmarkapp.com/developer/api/overview#error-codes]
							0 => no API error (it does not mean email delivery was successful)
		res.Message:		response message "OK" => email processed for delivery,
							check postmark account dashboard or use webhooks for actual delivery status
	*/

	return res, err
}
