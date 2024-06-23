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

	var weatherResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		log.Error(fmt.Sprintf("Error decoding weather data: %v", err))
		return WeatherResponse{}, err
	}

	log.Info(fmt.Sprintf("Successfully fetched weather data for city: %s", city))
	return weatherResponse, nil
}
