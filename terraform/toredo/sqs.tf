resource "aws_sqs_queue" "download_queue" {
  name                       = "${var.product_name}-download-queue"
  visibility_timeout_seconds = 43200
}