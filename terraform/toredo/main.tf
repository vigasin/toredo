provider "aws" {
  region = "us-west-1"
}

terraform {
  backend "s3" {
    encrypt = "true"
    bucket = "toredo-state"
    key = "toredo/terraform.tfstate"
    region = "us-west-1"
    dynamodb_table = "toredo-state-lock"
  }
}
