resource "aws_sqs_queue" "download_queue" {
  name                       = "${var.product_name}-download-queue"
  visibility_timeout_seconds = 43200
}

resource "aws_sqs_queue" "transfer_queue" {
  name                       = "${var.product_name}-transfer-queue"
  visibility_timeout_seconds = 43200
}