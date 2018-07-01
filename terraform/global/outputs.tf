output "lambda_bucket" {
  value = "${aws_s3_bucket.toredo_lambda.bucket}"
}