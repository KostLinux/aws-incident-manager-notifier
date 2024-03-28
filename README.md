# AWS Incident Manager Slack Notifier

This is a simple Slack notifier, which looks who's on-call this week and notifies them on Slack. It's designed to be run as a scheduled Lambda function.

## How it works

1. The Lambda function is triggered by a CloudWatch Event Rule Weekly every Monday
2. The function looks up the current week's on-call person in AWS Systems Manager Parameter Store
3. The function sends a message to the Slack channel of your choice
4. The function changes the on-call person before to the next person in the list


## Getting started

**1. Configure environment variables**

- `AWS_REGION`: The AWS region where the Lambda function is deployed
- `AWS_ACCESS_KEY_ID`: The AWS access key ID for Terraform
- `AWS_SECRET_ACCESS_KEY`: The AWS secret access key for Terraform
- `SLACK_WEBHOOK_URL`: The URL of the Slack Channel webhook to send messages to
- `SLACK_API_URL`: The Slack API URL to automatically change tag for the on-call person in slack
- `SLACK_API_TOKEN`: The Slack API token to authenticate with the Slack API
- `SLACK_SUBTEAM_ID`: The Slack subteam ID of the on-call team (e.g. `S123456`)
- `SLACK_SUBTEAM_NAME`: The Slack subteam name of the on-call team (e.g. `on-call`)
- `SSM_USERNAME_USER_ID`: This is the Slack ID for the person who is currently on-call. 

It's important to note that the prefix `SSM` must be in lowercase when used in the Incident Manager for proper parsing. 
For example, if the contact in the Incident Manager is `ssm_myname`, then the corresponding environment variable should be `SSM_MYNAME`. 
The value of this variable should be the Slack ID of the user. 
This way, the script can correctly identify and notify the on-call person.

So in short, the environment variable syntax is `"SSM_contact_unique_id" = "Slack User ID"`.

**2. Build your application to be Lambda compatible**

```
GOOS=linux GOARCH=amd64 go build -o bootstrap
```

**3. Use files from terraform folder to deploy lambda function**

```
terraform init -backend-config="bucket=your-bucket-name" -backend-config="key=terraform.tfstate" -backend-config="encrypt=true"
terraform plan
terraform apply
```

## License

This project is licensed under Apache License 2.0 - see the [LICENSE](LICENSE) file for details


## Author

* **KostLinux** - *Getting errors after error* - [KostLinux](https://github.com/KostLinux/)