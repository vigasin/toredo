output "downloader_user_name" {
  value = "${aws_iam_user.toredo_downloader.name}"
}

output "downloader_user_secret" {
  value = "${aws_iam_access_key.toredo_downloader.secret}"
}