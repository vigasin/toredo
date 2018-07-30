variable "manager_version" {
  default = "0.2"
}

resource "aws_lambda_function" "manager" {
  s3_bucket     = "${data.terraform_remote_state.global.lambda_bucket}"
  s3_key        = "manager/manager-v${var.manager_version}.zip"
  function_name = "HandleApiEvent"

  handler       = "manager"
  runtime       = "go1.x"

  role          = "${aws_iam_role.main_exec.arn}"
}

resource "aws_iam_role" "main_exec" {
  name               = "toredo_manager_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_iam_policy_document" "downloader_in_send" {
  statement {
    actions   = [
      "sqs:SendMessage"
    ]
    resources = [
      "${aws_sqs_queue.downloader_in.arn}",
    ]
  }
}

resource "aws_iam_policy" "downloader_in_send_lambda" {
  name   = "toredo-downloader_in-send-lambda"
  policy = "${data.aws_iam_policy_document.downloader_in_send.json}"
}

resource "aws_iam_role_policy_attachment" "downloader_in_send_lambda_att" {
  role       = "${aws_iam_role.main_exec.name}"
  policy_arn = "${aws_iam_policy.downloader_in_send_lambda.arn}"
}
