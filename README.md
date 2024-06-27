# Weather Lambda Function Demo

This repository contains a demo of a Lambda function for a weather app. The Lambda function is written in [Golang](https://golang.org/) and is designed to retrieve weather data from the third-party API, [Tomorrow.io](https://tomorrow.io/).

## Prerequisites

Before running the Lambda function, make sure you have the following:

- An AWS account
- AWS CLI installed and configured
- Golang installed
- Terraform installed

## Setup Terraform Backend

For best practices, we will store the Terraform state in an S3 bucket. Follow these steps to set up the backend:

1. **Create an S3 Bucket and DynamoDB Table** (only needed once):

   ```sh
   aws s3api create-bucket \
    --bucket <unique_bucket_name> \
    --region us-west-2 \
    --create-bucket-configuration LocationConstraint=us-west-2

   aws dynamodb create-table \
    --table-name terraform-lock \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST
   ```

2. **Update the `backend.tf` file in the `devops` directory**:
   ```hcl
   terraform {
     backend "s3" {
       bucket         = "unique_bucket_name>"
       key            = "weather-app/terraform.tfstate"
       region         = "us-west-2"
       dynamodb_table = "terraform-lock"
     }
   }
   ```

## Installation and Deployment

To install and deploy the Lambda function, follow these steps:

1. **Clone the Repository**:

   ```sh
   git clone https://github.com/SDBurt/go-weather-lambda.git
   cd go-weather-lambda
   ```

2. **Navigate to the Project Directory**:

   ```sh
   cd app
   ```

3. **Build and Zip the Executable for your Lambda Function**:

   ```sh
   make build
   ```

4. **Navigate to the `devops` Directory**:

   ```sh
   cd ../devops
   ```

5. **Initialize Terraform**:

   ```sh
   terraform init
   ```

6. **Apply the Terraform Configuration**:
   ```sh
   terraform apply -auto-approve
   ```
   This command will deploy all the necessary AWS resources including the IAM role, Lambda function, and DynamoDB table. Note the output values for `function_url`.

## Usage

To use the Lambda function, follow these steps:

1. **Make a Request to the Function URL**:
   ```sh
   curl "<function-url>?city=<city-name>"
   ```
   Replace `<function-url>` with the URL provided in the Terraform output and `<city-name>` with the desired city name (e.g., `NewYork`).

## Testing

To test the Lambda function, you can use the `curl` command as shown in the usage section. The function URL provided by Terraform will accept query parameters and return the weather data for the specified city.

Example:

```sh
curl "https://<your-lambda-url>?city=NewYork"
```

## Cleanup

To remove the Lambda function and associated resources once you are done, follow these steps:

1. **Destroy the Terraform-managed Infrastructure**:
   ```sh
   terraform destroy -auto-approve
   ```

## Project Structure

```plaintext
weather-app/
│
├── app/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── handler/
│   │   │   └── handler.go
│   │   ├── weather/
│   │   │   └── weather.go
│   │   ├── cache/
│   │   │   └── cache.go
│   │   ├── db/
│   │   │   └── db.go
│   │   └── log/
│   │       └── log.go
│   ├── go.mod
│   ├── go.sum
│   ├── .env
│   ├── Makefile
│   └── lambda-handler.zip
│
├── devops/
│   ├── main.tf
│   ├── backend.tf
│   ├── variables.tf
│   └── outputs.tf
│
└── README.md
```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
