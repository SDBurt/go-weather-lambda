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
