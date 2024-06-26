terraform {
  backend "s3" {
    bucket         = "sdburt-terraform-state-bucket"
    key            = "weather-app/terraform.tfstate"
    region         = "us-west-2"
    dynamodb_table = "terraform-lock"
  }
}