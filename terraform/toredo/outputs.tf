output "downloader_user_name" {
  value = "${aws_iam_user.toredo_downloader.name}"
}

output "downloader_user_id" {
  value = "${aws_iam_access_key.toredo_downloader.id}"
}

output "downloader_user_secret" {
  value = "${aws_iam_access_key.toredo_downloader.secret}"
}

output "transferer_user_name" {
  value = "${aws_iam_user.toredo_transferer.name}"
}

output "transferer_user_secret" {
  value = "${aws_iam_access_key.toredo_transferer.secret}"
}
