terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  s3_use_path_style           = true

  endpoints {
    s3  = "http://localstack:4566"
    sqs = "http://localstack:4566"
  }
}

resource "aws_s3_bucket" "outbox" {
  bucket = "outbox"
}

resource "aws_sqs_queue" "outbox" {
  name = "outbox"
}

resource "aws_s3_bucket_notification" "outbox" {
  bucket = aws_s3_bucket.outbox.id
  queue {
    queue_arn = aws_sqs_queue.outbox.arn
    events    = ["s3:ObjectCreated:*"]
  }
}
