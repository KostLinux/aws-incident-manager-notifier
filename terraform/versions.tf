
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.40.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.0.0"
    }
  }
  backend "s3" {}
}