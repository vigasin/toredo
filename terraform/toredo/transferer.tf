resource "aws_iam_user" "toredo_transferer" {
  name = "toredo_transferer"
  path = "/misc/"
}


resource "aws_iam_access_key" "toredo_transferer" {
  user = "${aws_iam_user.toredo_transferer.name}"
}

resource "aws_iam_user_policy" "transferer_policy" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage"
      ],
      "Effect": "Allow",
      "Resource": "${aws_sqs_queue.transferer_in.arn}"
    },
    {
      "Action": [
        "sqs:SendMessage"
      ],
      "Effect": "Allow",
      "Resource": "${aws_sqs_queue.transferer_out.arn}"
    },
    {
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage"
      ],
      "Effect": "Allow",
      "Resource": "${aws_sqs_queue.downloader_out.arn}"
    }
  ]
}
EOF
  user   = "${aws_iam_user.toredo_transferer.id}"
}

resource "aws_sqs_queue" "transferer_in" {
  name                       = "${var.product_name}-transferer-in"
  visibility_timeout_seconds = 43200
}

resource "aws_sqs_queue" "transferer_out" {
  name                       = "${var.product_name}-transferer-out"
  visibility_timeout_seconds = 43200
}