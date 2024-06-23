output "lambda_function_arn" {
  description = "The ARN of the Lambda function"
  value       = aws_lambda_function.weather_app.arn
}

output "function_url" {
  description = "The URL endpoint for the Lambda function"
  value = aws_lambda_function_url.weather_app_url.function_url
}
