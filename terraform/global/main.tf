provider "aws" {
  region = "us-west-1"
}

# Chicken/Egg problem here.
# Need to comment this initially
terraform {
  backend "s3" {
    encrypt        = "true"
    bucket         = "toredo-state"
    key            = "global/terraform.tfstate"
    region         = "us-west-1"
    dynamodb_table = "toredo-state-lock"
  }
}

resource "aws_dynamodb_table" "toredo_state_lock" {
  name           = "toredo-state-lock"
  hash_key       = "LockID"
  read_capacity  = 20
  write_capacity = 20

  attribute {
    name = "LockID"
    type = "S"
  }

  tags {
    Name = "DynamoDB Terraform State Lock Table"
  }
}

resource "aws_s3_bucket" "toredo_state" {
  bucket = "toredo-state"

  versioning {
    enabled = true
  }

  lifecycle {
    prevent_destroy = true
  }

  tags {
    Name = "S3 Remote Terraform State Store"
  }
}

resource "aws_s3_bucket" "toredo_lambda" {
  bucket = "toredo-lambda"

  lifecycle {
    prevent_destroy = true
  }

  tags {
    Name = "Bucket for toredo lambda packages"
  }
}

