resource "aws_iam_user" "toredo_downloader" {
  name = "toredo_downloader"
  path = "/misc/"
}


resource "aws_iam_access_key" "toredo_downloader" {
  user = "${aws_iam_user.toredo_downloader.name}"
}

resource "aws_iam_user_policy" "downloader_policy" {
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
      "Resource": "${aws_sqs_queue.download_queue.arn}"
    }
  ]
}
EOF
  user   = "${aws_iam_user.toredo_downloader.id}"
}