package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-redis/redis"
)

type RequestBody struct {
	Email string `json:"email"`
}

type QueueRequestBody struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func sendEmailToSQS(email string, code string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	queueUrl := os.Getenv("QUEUE_URL")

	if queueUrl == "" {
		return fmt.Errorf("unable to get QUEUE_URL")
	}

	request, err := json.Marshal(&QueueRequestBody{Email: email, Code: code})

	if err != nil {
		return fmt.Errorf("unable to marshall %v", err)
	}

	svc := sqs.NewFromConfig(cfg)
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(request)),
		QueueUrl:    aws.String(queueUrl),
	}

	response, err := svc.SendMessage(context.TODO(), input)

	if err != nil {
		return fmt.Errorf("unable to send message to queue %v", err)
	}

	fmt.Printf("Response %v", response)

	return nil
}

func generateCodeAndSendToRedis(email string) (string, error) {
	redisAddr, exists := os.LookupEnv("REDIS_CACHE")

	if !exists {
		return "", fmt.Errorf("unable to get REDIS_CACHE")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	rand.Seed(time.Now().UnixNano())

	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	err := rdb.Set(email, code, 0).Err()
	if err != nil {
		return "", err
	}

	val, err := rdb.Get(email).Result()
	if err != nil {
		return "", err
	}
	fmt.Println("key", val)

	return code, nil
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var record RequestBody
	err := json.Unmarshal([]byte(event.Body), &record)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error occurred marshalling request body",
		}, err
	}

	code, err := generateCodeAndSendToRedis(record.Email)

	if err != nil {
		fmt.Printf("error with redis %v", err)
	}

	err = sendEmailToSQS(record.Email, code)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error occurred queueing email",
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Email queued",
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
