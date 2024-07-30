package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
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
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("unable to load SDK config, %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	listObjectsParams := &s3.ListObjectsV2Input{
		Bucket: aws.String(targetBucketName),
		Prefix: aws.String(targetFolderPath + "/" + folderName + "/"),
	}
	listObjectsOutput, err := s3Client.ListObjectsV2(ctx, listObjectsParams)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to list objects, %w", err)
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, item := range listObjectsOutput.Contents {
		getObjectParams := &s3.GetObjectInput{
			Bucket: aws.String(targetBucketName),
			Key:    aws.String(*item.Key),
		}
		getObjectOutput, err := s3Client.GetObject(ctx, getObjectParams)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to get object %s, %w", *item.Key, err)
		}
		defer getObjectOutput.Body.Close()

		zipFileWriter, err := zipWriter.Create(*item.Key)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to create zip entry for %s, %w", *item.Key, err)
		}

		_, err = io.Copy(zipFileWriter, getObjectOutput.Body)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to write zip entry for %s, %w", *item.Key, err)
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to close zip writer, %w", err)
	}

	zipKey := fmt.Sprintf("%s/%s.zip", targetFolderPath, folderName)
	putObjectParams := &s3.PutObjectInput{
		Bucket: aws.String(targetBucketName),
		Key:    aws.String(zipKey),
		Body:   bytes.NewReader(buf.Bytes()),
	}

	_, err = s3Client.PutObject(ctx, putObjectParams)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("failed to upload zip file, %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Folder downloaded, zipped, and uploaded to S3",
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
