# Building a Weather Application with AWS Lambda and API Gateway

In today's world, serverless computing and Infrastructure as Code (IaC) have become game-changers for developers. This blog post will guide you through building a weather application using AWS Lambda and API Gateway, with all infrastructure managed by Terraform. This application fetches weather data from [Tomorrow.io](https://tomorrow.io) for a given city and stores it in DynamoDB. The entire project is written in Go, a powerful and efficient programming language.

The code for this project can be found in my github repo [here](https://github.com/SDBurt/go-weather-lambda)

## Prerequisites

Before we start, ensure you have the following prerequisites:

- An AWS account
- AWS CLI installed and configured
- Golang installed
- Terraform installed

## Overview of the Project

The weather application leverages the following AWS services:

- **AWS Lambda**: Executes the application code in response to HTTP requests.
- **API Gateway**: Provides a RESTful API endpoint for the Lambda function.
- **DynamoDB**: Stores the fetched weather data.
- **IAM Roles and Policies**: Manages permissions for accessing AWS resources.

## Project Structure

The project is organized as follows:

```plaintext
weather-app/
│
├── app/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── handler/
│   │   │   └── handler.go
│   │   ├── weather/
│   │   │   └── weather.go
│   │   ├── cache/
│   │   │   └── cache.go
│   │   ├── db/
│   │   │   └── db.go
│   │   └── log/
│   │       └── log.go
│   ├── go.mod
│   ├── go.sum
│   ├── .env
│   ├── Makefile
│   └── lambda-handler.zip
│
├── devops/
│   ├── main.tf
│   ├── backend.tf
│   ├── variables.tf
│   └── outputs.tf
│
└── README.md
```

## Setting Up Terraform Backend

Using Terraform to manage infrastructure state remotely is a best practice. We will use an S3 bucket and a DynamoDB table for state locking and consistency.

### Step 1: Create an S3 Bucket and DynamoDB Table

Execute the following commands to set up the backend:

```sh
aws s3api create-bucket --bucket <unique_bucket_name> --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2
aws dynamodb create-table --table-name terraform-lock --attribute-definitions AttributeName=LockID,AttributeType=S --key-schema AttributeName=LockID,KeyType=HASH --billing-mode PAY_PER_REQUEST
```

This code creates a bucket to store the state and a dynamodb table to handle locking the state for deployments.

### Step 2: Update the `backend.tf` File

Update the `backend.tf` file with your S3 bucket name and DynamoDB table name:

```hcl
terraform {
  backend "s3" {
    bucket         = "<unique_bucket_name>"
    key            = "weather-app/terraform.tfstate"
    region         = "us-west-2"
    dynamodb_table = "terraform-lock"
  }
}
```

## Setting Up the Terraform Configuration

We'll create the necessary AWS resources using Terraform. Below is the configuration for our infrastructure.

### main.tf

```hcl
provider "aws" {
  region = "us-west-2"
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "lambda_execution_role" {
  name = "lambda-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        },
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_execution_policy_attachment" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_policy" "lambda_dynamodb_policy" {
  name        = "lambda-dynamodb-policy"
  description = "Policy for Lambda to access DynamoDB"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:UpdateItem"
        ],
        Resource = "arn:aws:dynamodb:us-west-2:${data.aws_caller_identity.current.account_id}:table/${var.DB_TABLE_NAME}"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_dynamodb_policy_attachment" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.lambda_dynamodb_policy.arn
}

resource "aws_dynamodb_table" "weather_data" {
  name         = var.DB_TABLE_NAME
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "City"

  attribute {
    name = "City"
    type = "S"
  }
}

resource "aws_lambda_function" "weather_app" {
  function_name = "weather-app"
  role          = aws_iam_role.lambda_execution_role.arn
  handler       = "main"
  runtime       = "provided.al2023"
  filename      = "${path.module}/../app/lambda-handler.zip"
  architectures = ["arm64"]

  environment {
    variables = {
      DB_TABLE_NAME = aws_dynamodb_table.weather_data.name
      API_KEY       = var.WEATHER_API_KEY
      VERSION       = var.VERSION
    }
  }
}

resource "aws_apigatewayv2_api" "api" {
  name          = "weather-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "integration" {
  api_id           = aws_apigatewayv2_api.api.id
  integration_type = "AWS_PROXY"
  integration_uri  = aws_lambda_function.weather_app.arn
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "route" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "GET /weather"
  target    = "integrations/${aws_apigatewayv2_integration.integration.id}"
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.weather_app.arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*"
}

resource "aws_apigatewayv2_stage" "stage" {
  api_id      = aws_apigatewayv2_api.api.id
  name        = "$default"
  auto_deploy = true
}
```

### variables.tf

```hcl
variable "WEATHER_API_KEY" {
  description = "API key for the Tomorrow.io API"
  type        = string
}

variable "DB_TABLE_NAME" {
  description = "Name of the DynamoDB table to store weather data"
  type        = string
  default     = "weather-data"
}

variable "VERSION" {
  description = "Version of the Lambda function"
  type        = string
}
```

### outputs.tf

```hcl
output "api_endpoint" {
  description = "The URL endpoint for the API"
  value       = aws_apigatewayv2_stage.stage.invoke_url
}
```

### backend.tf

```hcl
terraform {
  backend "s3" {
    bucket = "your-terraform-state-bucket"
    key    = "path/to/your/terraform.tfstate"
    region = "us-west-2"
  }
}
```

## Writing the Lambda Function

Next, let's write the Lambda function in Go. This function handles requests from API Gateway, fetches weather data, caches the results, and stores the data in DynamoDB.

### handler.go

```go
package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"weather-lambda/internal/cache"
	"weather-lambda/internal/db"
	"weather-lambda/internal/log"
	"weather-lambda/internal/weather"

	"github.com/aws/aws-lambda-go/events"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	city := request.QueryStringParameters["city"]

    // Sanitize city parameter
    sanitizedCity := url.QueryEscape(city)

	// Validate city
	if sanitizedCity == "" {
		log.Error("City parameter is required")
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	// Check cache first
	if cachedData, found := cache.GetCache(sanitizedCity); found {
		log.Info(fmt.Sprintf("Returning cached data for city: %s", sanitizedCity))
		return buildResponse(cachedData)
	}

	// Fetch weather data
	weatherResponse, err := weather.FetchWeather(sanitizedCity)
	if err != nil {
		log.Error(fmt.Sprintf("Error fetching weather data: %v", err))
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	weatherData := weatherResponse.Data.Values

	// Save to DynamoDB
	dbData := db.WeatherData{
		City:        sanitizedCity,
		Temperature: weatherData.Temperature,
		Humidity:    weatherData.Humidity,
	}

	if err := db.SaveWeatherData(dbData); err != nil {
		log.Error(fmt.Sprintf("Error saving weather data to DynamoDB: %v", err))
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Cache the response
	cache.SetCache(sanitizedCity, dbData)

	log.Info(fmt.Sprintf("Returning new data for city: %s", sanitizedCity))
	return buildResponse(dbData)
}

func buildResponse(data interface{}) (events.APIGatewayProxyResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Error(fmt.Sprintf("Error marshalling response data: %v", err))
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}
```

### main.go

```go
package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    "weather-lambda/internal/handler"
)

func main() {
    lambda.Start(handler.HandleRequest)
}
```

### cache.go

```go
package cache

import (
	"fmt"
	"time"
	"weather-lambda/internal/log"

	"github.com/patrickmn/go-cache"
)

var c = cache.New(5*time.Minute, 10*time.Minute)

func SetCache(key string, value interface{}) {
	log.Info(fmt.Sprintf("Setting cache for key: %s", key))
	c.Set(key, value, cache.DefaultExpiration)
}

func GetCache(key string) (interface{}, bool) {
	data, found := c.Get(key)
	if found {
		log.Info(fmt.Sprintf("Cache hit for key: %s", key))
	} else {
		log.Info(fmt.Sprintf("Cache miss for key: %s", key))
	}
	return data, found
}
```

### weather.go

```go
package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"weather-lambda/internal/log"
)

type WeatherDataValues struct {
	CloudBase                interface{} `json:"cloudBase"`
	CloudCeiling             interface{} `json:"cloudCeiling"`
	CloudCover               int         `json:"cloudCover"`
	DewPoint                 float64     `json:"dewPoint"`
	FreezingRainIntensity    int         `json:"freezingRainIntensity"`
	Humidity                 int         `json:"humidity"`
	PrecipitationProbability int         `json:"precipitationProbability"`
	PressureSurfaceLevel     float64     `json:"pressureSurfaceLevel"`
	RainIntensity            int         `json:"rainIntensity"`
	SleetIntensity           int         `json:"sleetIntensity"`
	SnowIntensity            int         `json:"snowIntensity"`
	Temperature              float64     `json:"temperature"`
	TemperatureApparent      float64     `json:"temperatureApparent"`
	UVHealthConcern          int         `json:"uvHealthConcern"`
	UVIndex                  int         `json:"uvIndex"`
	Visibility               float64     `json:"visibility"`
	WeatherCode              int         `json:"weatherCode"`
	WindDirection            float64     `json:"windDirection"`
	WindGust                 float64     `json:"windGust"`
	WindSpeed                float64     `json:"windSpeed"`
}

type WeatherData struct {
	Time   string            `json:"time"`
	Values WeatherDataValues `json:"values"`
}

type WeatherLocation struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Name string  `json:"name"`
	Type string  `json:"type"`
}

type WeatherResponse struct {
	Data     WeatherData     `json:"data"`
	Location WeatherLocation `json:"location"`
}

func FetchWeather(city string) (WeatherResponse, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")

	if apiKey == "" {
		return WeatherResponse{}, fmt.Errorf("WEATHER_API_KEY is required")
	}

	url := fmt.Sprintf("https://api.tomorrow.io/v4/weather/realtime?location=%s&apikey=%s", city, apiKey)

	log.Info(fmt.Sprintf("Fetching weather data for city: %s", city))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf("Error making HTTP request: %v", err))
		return WeatherResponse{}, err
	}
	defer resp.Body.Close()

	log.Info(fmt.Sprintf("Received response with status code: %d", resp.StatusCode))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WeatherResponse{}, fmt.Errorf("received response with status code: %d", resp.StatusCode)
	}

	log.Info(fmt.Sprintf("Response: %+v", resp))

	var weatherResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		log.Error(fmt.Sprintf("Error decoding weather data: %v", err))
		return WeatherResponse{}, err
	}

	log.Info(fmt.Sprintf("Successfully fetched weather data for city: %s", city))
	return weatherResponse, nil
}
```

### log.go

```go
package log

import (
	"log"
	"os"
)

var (
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func Info(msg string) {
	infoLogger.Println(msg)
}

func Error(msg string) {
	errorLogger.Println(msg)
}
```

### db.go

```go
package db

import (
	"fmt"
	"os"
	"weather-lambda/internal/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type WeatherData struct {
	City        string  `json:"City"`
	Temperature float64 `json:"Temperature"`
	Humidity    int     `json:"Humidity"`
}

func SaveWeatherData(data WeatherData) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Error(fmt.Sprintf("Error marshalling weather data: %v", err))
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Error(fmt.Sprintf("Error saving weather data to DynamoDB: %v", err))
		return err
	}

	log.Info(fmt.Sprintf("Successfully saved weather data for city: %s", data.City))
	return nil
}
```

### .env (example)

```plaintext
WEATHER_API_KEY=<your_tomorrow_io_api_key>
DB_TABLE_NAME=weather-data
```

## Deploying the Lambda Function

Compile the Go code and create a zip file to deploy the Lambda function.
For this particular project, we are building and deploying to the arm64 architecture.

```sh
GOOS=linux GOARCH=arm64 go build -o main main.go
zip lambda-handler.zip main
```

## Applying the Terraform Configuration

Initialize and apply the Terraform configuration to deploy the resources.

```sh
terraform init
terraform apply
```

## Testing the API

Once the resources are deployed, use the output `api_endpoint` to make requests to your API.

```sh
curl -X GET "<api_endpoint>/weather?city=london"
```

## Cleaning Up

To remove the Lambda function and associated resources, destroy the Terraform-managed infrastructure.

```sh
terraform destroy -auto-approve
```

## Conclusion

This blog post walked you through setting up a weather application using AWS Lambda and API Gateway with Terraform. By following these steps, you can build a scalable, serverless application with Go and manage it using Infrastructure as Code.

Feel free to explore and modify the project. Contributions are welcome! If you encounter any issues or have suggestions for improvements, please open an issue or submit a pull request on the project's GitHub repository.

Happy coding!
