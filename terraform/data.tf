data "archive_file" "slack_ssm_notifier_zip" {
  type        = "zip"
  source_dir  = "${path.module}"
  output_path = "${path.module}/oncall-notify.zip"
}