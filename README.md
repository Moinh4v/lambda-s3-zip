# S3 Folder Zipper Lambda Function

Lambda function zips the contents of a specified folder in an S3 bucket and uploads the resulting zip file back to the same or another S3 bucket.

## Purpose

The function:
 1. Lists all objects in a specified folder in an S3 bucket.
 2. Downloads each object.
 3. Zips the downloaded objects into a single zip file.
 4. Uploads the zip file back to S3.

## Prerequisites

- AWS account with access to S3.
- An S3 bucket to read from and write to.
- AWS Lambda setup with the appropriate IAM permissions to read from and write to S3.

## Environment Variables

The following environment variables must be set for the Lambda function:

- `TARGET_BUCKET_NAME`: The name of the S3 bucket containing the folder to be zipped.
- `TARGET_FOLDER_PATH`: The path to the folder within the S3 bucket to be zipped (excluding the folder name).

## Request

The function expects an `APIGatewayProxyRequest` with the following path parameter:

- `folder-name`: The name of the folder within the `TARGET_FOLDER_PATH` to be zipped.

## Setup and Deployment

### 1. Create IAM Role

Create an IAM role with the necessary permissions to access S3. Attach the following policy to the role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::YOUR_BUCKET_NAME",
        "arn:aws:s3:::YOUR_BUCKET_NAME/*"
      ]
    }
  ]
}
```

### 2. Create Lambda Function

1. Go to the AWS Lambda console and create a new function.
2. Set the runtime to Go.
3. Upload the `main.go` file with the code provided.
4. Set the environment variables `TARGET_BUCKET_NAME` and `TARGET_FOLDER_PATH`.
5. Attach the IAM role created in step 1 to the Lambda function.

### 3. Create API Gateway

1. Create a new API Gateway.
2. Create a new resource and method (e.g., GET) that triggers the Lambda function.
3. Deploy the API.

## Usage

Invoke the Lambda function via the API Gateway with the folder name as a path parameter. For example:

```bash
GET https://your-api-id.execute-api.region.amazonaws.com/prod/{folder-name}
```


This request zips the contents of `my-folder` within the specified `TARGET_FOLDER_PATH` and uploads the zip file to the same S3 bucket.

## Testing

You can test the function locally by using the provided test case in `main_test.go`. Run the test with:

```bash
go test -v
```

### Expected Output

The test should list objects in the specified S3 folder, zip them, and upload the resulting zip file back to S3. The test log should display the size of the zip file and a preview of its contents.