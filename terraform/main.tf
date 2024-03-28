variable "oncall_slack_reminder" {
  type = list(object({
    name  = string
    value = string
  }))
  description = "The environment variables for the Lambda function"
  default = [
    {
        name = "AWS_REGION"
        value = "us-east-1"
    },
    {
        name = "SLACK_API_URL"
        value = "https://slack.com/api/usergroups.users.update"
    },
    {
        name = "SLACK_WEBHOOK_URL"
        value = "https://hooks.slack.com/services/FOO/BAR/BAZ"
    },
    {
        name = "SLACK_API_TOKEN"
        value = "xyz"
    },
    {
        name = "SLACK_SUBTEAM_ID"
        value = "S123456"
    },
    {
        name = "SLACK_SUBTEAM_NAME"
        value = "support"
    },
    {
        name = "SSM_JOHN_DOE_USER_ID"
        value = "FOOBARXYZ12"
    },
    {
        name = "SSM_JANE_DOE_USER_ID"
        value = "FOOBARXYZ34"
    }
  ]
  
}

# OnCall Notifier IAM
resource "aws_iam_role" "ssm_incident_manager_role" {
  name               = "SSMIncidentManagerReadOnlyRole"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
resource "aws_iam_policy" "ssm_incident_manager_read_only" {
  name        = "SSMIncidentManagerReadOnly"
  description = "Read-only access to SSM Incident Manager"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm-contacts:Read*",
        "ssm-contacts:List*",
        "ssm-contacts:Get*",
        "ssm-contacts:Describe*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ssm_incident_manager_role_policy_attachment" {
  role       = aws_iam_role.ssm_incident_manager_role.name
  policy_arn = aws_iam_policy.ssm_incident_manager_read_only.arn
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution_role_policy_attachment" {
  role       = aws_iam_role.ssm_incident_manager_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# OnCall Notifier Lambda
resource "aws_lambda_function" "slack_ssm_notifier" {
    function_name = "slack-oncall_notifier"
    description   = "Notifier admins who are oncall"
    handler       = "bootstrap"
    runtime       = "provided.al2023"
    role          = aws_iam_role.ssm_incident_manager_role.arn
    filename      = data.archive_file.slack_ssm_notifier_zip.output_path
    source_code_hash = filebase64sha256(data.archive_file.slack_ssm_notifier_zip.output_path)
    tracing_config {
        mode = "Active"
    }
    tags = {
        Name = "slack-oncall_notifier"
    }

    environment {
        variables = { for obj in var.oncall_slack_reminder : obj.name => obj.value }
    }
}

# Create the EventBridge rule that notifies every Monday
resource "aws_cloudwatch_event_rule" "lambda_event_notifier" {
    name = "eventbridge-trigger-slack-oncall-lambda"
    description = "scheduled every Monday"
    schedule_expression = "cron(0 8 ? * MON *)"
}

resource "aws_cloudwatch_event_target" "lambda_target_notifier" {
    arn = aws_lambda_function.slack_ssm_notifier.arn
    rule = aws_cloudwatch_event_rule.lambda_event_notifier.name
}

resource "aws_lambda_permission" "allow_execution_from_cloudwatch" {
    statement_id = "AllowExecutionFromCloudWatch"
    action = "lambda:InvokeFunction"
    function_name = aws_lambda_function.slack_ssm_notifier.arn
    principal = "events.amazonaws.com"
    source_arn = aws_cloudwatch_event_rule.lambda_event_notifier.arn
}
