package db

import (
	"fmt"
	"weather-lambda/internal/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type WeatherData struct {
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
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

	log.Info(fmt.Sprintf("Successfully saved weather data for name: %s", data.Name))
	return nil
}
