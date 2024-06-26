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
