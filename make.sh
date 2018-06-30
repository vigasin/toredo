#!/bin/bash

function publish_downloader()
{
    aws_user=$(cd terraform/toredo; terraform output downloader_user_name)
    aws_secret=$(cd terraform/toredo; terraform output downloader_user_secret)
    echo $aws_user $aws_secret
}

function publish_transferer()
{
    aws_user=$(cd ../terraform/toredo; terraform output transferer_user_name)
    aws_secret=$(cd ../terraform/toredo; terraform output transferer_user_secret)

    echo $aws_user $aws_secret
}