terraform {
  backend "s3" {
    bucket         = "weather-app-state-bucket"
    key            = "weather-app/terraform.tfstate"
    region         = "us-west-2"
    dynamodb_table = "weather-app-terraform-lock"
  }
}
