package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    "weather-lambda/internal/handler"
)

func main() {
    lambda.Start(handler.HandleRequest)
}

