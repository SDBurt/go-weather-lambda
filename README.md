
# Weather Lambda Function Demo

This repository contains a demo of a Lambda function for a weather app. The Lambda function is written in [Golang](https://golang.org/) and is designed to retrieve weather data from the third-party API, [Tomorrow.io](https://tomorrow.io/).

## Prerequisites

Before running the Lambda function, make sure you have the following:

- An AWS account
- AWS CLI installed and configured
- Golang installed

## Installation

To install and deploy the Lambda function, follow these steps:

1. Clone this repository to your local machine.
2. Navigate to the project directory: `cd weather-app`.
3. Build and zip the executable for your Lambda function: `make build`.
4. Upload the deployment package to AWS Lambda using the AWS CLI:
   ```sh
   aws lambda create-function \
   --function-name weather-app \
   --runtime provided.al2023 \
   --handler bootstrap \
   --zip-file fileb://deployment.zip \
   --architecture arm64
   ```
5. Run `make cleanup` to clean up the zip file

## Usage

To use the Lambda function, follow these steps:

1. Invoke the function using the AWS CLI:
   ```sh
   aws lambda invoke --function-name weather-app --payload '{"city": "Seattle"}' output.json
   ```
2. Check the `output.json` file for the weather data.

## Cleanup

Remove the lambda function once you are done by the following these steps:

1. If you don't know the function ARN, get it using the AWS CLI:
   ```sh
   aws lambda get-function --function-name weather-app
   ```
2. Invoke the delete function command using the AWS CLI:
   ```sh
   aws lambda delete-function --function-name arn:aws:lambda:<region>:<account-id>:function:weather-app
   ```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
