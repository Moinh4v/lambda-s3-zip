package main

import (
	"context"
	"fmt"

	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joho/godotenv"
)

func TestHandleRequest(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Fatalf("Error loading .env file")
	}

	request := events.APIGatewayProxyRequest{
		PathParameters: map[string]string{
			"folder-name": "2024-02-23",
		},
	}

	response, err := HandleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	fmt.Println("Response:", response)
}
