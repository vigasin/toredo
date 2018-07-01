terraform {
  backend "s3" {
    encrypt        = "true"
    bucket         = "toredo-state"
    key            = "toredo/terraform.tfstate"
    region         = "us-west-1"
    dynamodb_table = "toredo-state-lock"
  }
}

provider "aws" {
  region = "us-west-1"
}

data "terraform_remote_state" "global" {
  backend = "s3"

  config {
    bucket = "toredo-state"
    key    = "global/terraform.tfstate"
    region = "us-west-1"
  }
}