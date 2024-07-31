package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	folderName := request.PathParameters["folder-name"]
	targetBucketName := os.Getenv("TARGET_BUCKET_NAME")
	targetFolderPath := os.Getenv("TARGET_FOLDER_PATH")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading SDK config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("unable to load SDK config, %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// List objects in the specified folder
	listObjectsParams := &s3.ListObjectsV2Input{
		Bucket: aws.String(targetBucketName),
		Prefix: aws.String(folderName),
	}
	listObjectsOutput, err := s3Client.ListObjectsV2(ctx, listObjectsParams)
	if err != nil {
		log.Printf("Error listing objects: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to list objects, %w", err)
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, item := range listObjectsOutput.Contents {
		if *item.Key == fmt.Sprintf("%s/%s/", targetFolderPath, folderName) {
			continue // Skip directory entries
		}
		
		log.Printf("Processing file: %s", *item.Key)
		getObjectParams := &s3.GetObjectInput{
			Bucket: aws.String(targetBucketName),
			Key:    aws.String(*item.Key),
		}
		getObjectOutput, err := s3Client.GetObject(ctx, getObjectParams)
		if err != nil {
			log.Printf("Error getting object %s: %v", *item.Key, err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to get object %s, %w", *item.Key, err)
		}
		defer getObjectOutput.Body.Close()

		fileBuf := new(bytes.Buffer)
		n, err := io.Copy(fileBuf, getObjectOutput.Body)
		if err != nil {
			log.Printf("Error reading object %s: %v", *item.Key, err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to read object %s, %w", *item.Key, err)
		}
		log.Printf("Read %d bytes from file %s", n, *item.Key)

		// Removed the check for zero-length files
		relativePath := strings.TrimPrefix(*item.Key, fmt.Sprintf("%s/", targetFolderPath))
		log.Printf("Adding file to zip: %s", relativePath)
		zipFileWriter, err := zipWriter.Create(relativePath)
		if err != nil {
			log.Printf("Error creating zip entry for %s: %v", *item.Key, err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to create zip entry for %s, %w", *item.Key, err)
		}

		n, err = io.Copy(zipFileWriter, fileBuf)
		if err != nil {
			log.Printf("Error writing zip entry for %s: %v", *item.Key, err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to write zip entry for %s, %w", *item.Key, err)
		}
		log.Printf("Wrote %d bytes for file %s to zip", n, *item.Key)
	}

	err = zipWriter.Close()
	if err != nil {
		log.Printf("Error closing zip writer: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to close zip writer, %w", err)
	}

	totalSize := buf.Len()
	log.Printf("Total zip size: %d bytes", totalSize)

	zipKey := fmt.Sprintf("%s/%s.zip", targetFolderPath, folderName)
	putObjectParams := &s3.PutObjectInput{
		Bucket: aws.String(targetBucketName),
		Key:    aws.String(zipKey),
		Body:   bytes.NewReader(buf.Bytes()),
	}

	_, err = s3Client.PutObject(ctx, putObjectParams)
	if err != nil {
		log.Printf("Error uploading zip file: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to upload zip file, %w", err)
	}

	log.Printf("Successfully uploaded zip file to %s/%s.zip", targetBucketName, zipKey)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Folder downloaded, zipped, and uploaded to S3",
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
