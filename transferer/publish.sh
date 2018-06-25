#!/bin/bash

aws_user=$(cd ../terraform/toredo; terraform output transferer_user_name)
aws_secret=$(cd ../terraform/toredo; terraform output transferer_user_secret)

echo $aws_user $aws_secret
