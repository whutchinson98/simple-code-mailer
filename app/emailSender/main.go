package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "sender@example.com"

	// The subject line for the email.
	Subject = "Example Subject"

	// The character encoding for the email.
	CharSet = "UTF-8"
)

func sendEmail(recipient string, code string) error {
	htmlBody := fmt.Sprintf("<html><h1>Auth Code</h1><p>Your auth code is: %s</p></html>", code)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))

	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	svc := ses.NewFromConfig(cfg)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				recipient,
			},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(htmlBody),
				},
			},
			Subject: &types.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
	}

	_, err = svc.SendEmail(context.TODO(), input)

	if err != nil {
		return err
	}

	fmt.Println("Email Sent to address: " + recipient)

	return nil

}

type RecordStruct struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func HandleRequest(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	batchItemFailures := make([]events.SQSBatchItemFailure, 0)

	for _, e := range event.Records {
		var record RecordStruct
		json.Unmarshal([]byte(e.Body), &record)

		fmt.Printf("Sending email for %v %v\n", record.Email, record.Code)

		err := sendEmail(record.Email, record.Code)

		if err != nil {
			batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.Email})
		}
	}

	return events.SQSEventResponse{
		BatchItemFailures: batchItemFailures,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
