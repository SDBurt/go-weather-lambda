---
title: Golang Weather Aggregator Lambda Function
date: "2024-06-18"
draft: true
description: How to Build a Professional, Modular, and Scalable Golang Lambda Function
---

# Building a Professional, Modular, and Scalable Golang Lambda Function

In this blog post, we'll go beyond the typical "Hello World" example and build a professional, modular, and scalable AWS Lambda function using Go. We'll cover how to access other AWS services and implement caching for improved performance. This tutorial is designed to provide substantial value to developers looking to learn and apply Go to AWS Lambda in a real-world scenario.

## Prerequisites

- Basic knowledge of Golang
- AWS account with necessary permissions
- AWS CLI installed and configured
- Go environment set up

## Scenario: Weather Data Aggregator

We'll create a Lambda function that aggregates weather data from an external API and stores it in DynamoDB. The function will also cache the response to reduce external API calls, improving efficiency.

### Project Structure

We'll organize our project into the following structure for modularity and scalability:

```
weather-lambda/
├── cmd/
│   └── main.go
├── internal/
│   ├── handler/
│   │   └── handler.go
│   ├── weather/
│   │   └── weather.go
│   ├── cache/
│   │   └── cache.go
│   ├── db/
│   │   └── db.go
│   ├── log/
│   │   └── log.go
├── go.mod
└── go.sum
```

### Setting Up the Project

Initialize a new Go module:

```sh
go mod init weather-lambda
```

### Implementing the Weather API Client

Create `internal/weather/weather.go`:

```go
package weather

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "weather-lambda/internal/log"
)

type WeatherData struct {
    Temperature string `json:"temperature"`
    Humidity    string `json:"humidity"`
    Description string `json:"description"`
}

func FetchWeather(city string) (*WeatherData, error) {
    apiKey := os.Getenv("WEATHER_API_KEY")
    url := fmt.Sprintf("https://api.example.com/weather?city=%s&apikey=%s", city, apiKey)

    log.Info(fmt.Sprintf("Fetching weather data for city: %s", city))

    resp, err := http.Get(url)
    if err != nil {
        log.Error(fmt.Sprintf("Error fetching weather data: %v", err))
        return nil, err
    }
    defer resp.Body.Close()

    var weatherData WeatherData
    if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
        log.Error(fmt.Sprintf("Error decoding weather data: %v", err))
        return nil, err
    }

    log.Info(fmt.Sprintf("Successfully fetched weather data for city: %s", city))
    return &weatherData, nil
}
```

### Implementing the Cache Layer

Create `internal/cache/cache.go`:

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

### Implementing the Database Layer

Create `internal/db/db.go`:

```go
package db

import (
    "weather-lambda/internal/log"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type WeatherData struct {
    City        string `json:"city"`
    Temperature string `json:"temperature"`
    Humidity    string `json:"humidity"`
    Description string `json:"description"`
}

func SaveWeatherData(data WeatherData) error {
    sess := session.Must(session.NewSession())
    svc := dynamodb.New(sess)

    av, err := dynamodbattribute.MarshalMap(data)
    if err != nil {
        log.Error(fmt.Sprintf("Error marshalling weather data: %v", err))
        return err
    }

    input := &dynamodb.PutItemInput{
        Item:      av,
        TableName: aws.String("WeatherData"),
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

### Implementing the Logging Layer

Create `internal/log/log.go`:

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

### Implementing the Lambda Handler

Create `internal/handler/handler.go`:

```go
package handler

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "weather-lambda/internal/cache"
    "weather-lambda/internal/db"
    "weather-lambda/internal/log"
    "weather-lambda/internal/weather"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    city := request.QueryStringParameters["city"]

    // Check cache first
    if cachedData, found := cache.GetCache(city); found {
        log.Info(fmt.Sprintf("Returning cached data for city: %s", city))
        return buildResponse(cachedData)
    }

    // Fetch weather data
    weatherData, err := weather.FetchWeather(city)
    if err != nil {
        log.Error(fmt.Sprintf("Error fetching weather data: %v", err))
        return events.APIGatewayProxyResponse{StatusCode: 500}, err
    }

    // Save to DynamoDB
    dbData := db.WeatherData{
        City:        city,
        Temperature: weatherData.Temperature,
        Humidity:    weatherData.Humidity,
        Description: weatherData.Description,
    }
    if err := db.SaveWeatherData(dbData); err != nil {
        log.Error(fmt.Sprintf("Error saving weather data to DynamoDB: %v", err))
        return events.APIGatewayProxyResponse{StatusCode: 500}, err
    }

    // Cache the response
    cache.SetCache(city, dbData)

    log.Info(fmt.Sprintf("Returning new data for city: %s", city))
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

func main() {
    lambda.Start(HandleRequest)
}
```

### Deploying the Lambda Function

1. **Build the Lambda binary**:

   ```sh
   GOOS=linux GOARCH=amd64 go build -o main weather-lambda/cmd/main.go
   ```

2. **Create a deployment package**:

   ```sh
   zip deployment.zip main
   ```

3. **Deploy using AWS CLI**:

   ```sh
   aws lambda create-function --function-name WeatherAggregator     --zip-file fileb://deployment.zip --handler main     --runtime go1.x --role arn:aws:iam::<your-account-id>:role/<your-lambda-role>
   ```

### Conclusion

In this blog post, we demonstrated how to build a professional, modular, and scalable Golang Lambda function. We covered how to access other AWS services and implement caching for performance improvements. By structuring our code in a modular way, we can easily extend and maintain the function, making it suitable for real-world applications.

Feel free to clone the repository [here](#) and try it out yourself!

### References

- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)
- [AWS SDK for Go](https://aws.amazon.com/sdk-for-go/)
- [Go Cache Library](https://github.com/patrickmn/go-cache)
