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

	// Validate city
	if city == "" {
		log.Error("City parameter is required")
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	// Check cache first
	if cachedData, found := cache.GetCache(city); found {
		log.Info(fmt.Sprintf("Returning cached data for city: %s", city))
		return buildResponse(cachedData)
	}

	// Fetch weather data
	weatherResponse, err := weather.FetchWeather(city)
	if err != nil {
		log.Error(fmt.Sprintf("Error fetching weather data: %v", err))
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	weatherData := weatherResponse.Data.Values
	weatherLocation := weatherResponse.Location

	log.Info(fmt.Sprintf("WeatherData for city: %s, %+v", city, weatherData))
	log.Info(fmt.Sprintf("WeatherLocation for city: %s, %+v", city, weatherLocation))

	// Save to DynamoDB
	dbData := db.WeatherData{
		City:        city,
		Temperature: weatherData.Temperature,
		Humidity:    weatherData.Humidity,
	}

	log.Info(fmt.Sprintf("Saving %+v to DynamoDB", dbData))

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
