terraform {
  backend "s3" {
    bucket = "otanikotani-goaswtf"
    key = "filename.tfstate"
    region = "us-east-1"
    encrypt = true
  }
}

# This is required to get the AWS region via ${data.aws_region.current}.
provider "aws" {
  region  = "us-east-1"
}
# Define a Lambda function.
#
# The handler is the name of the executable for go1.x runtime.
resource "aws_lambda_function" "goaswtf" {
  function_name    = "goaswtf"
  filename         = "goaswtf.zip"
  handler          = "goaswtf"
  source_code_hash = filebase64sha256("goaswtf.zip")
  role             = aws_iam_role.goaswtf.arn
  runtime          = "go1.x"
  memory_size      = 128
  timeout          = 1
}

# A Lambda function may access to other AWS resources such as S3 bucket. So an
# IAM role needs to be defined. This example does not access to
# any resource, so the role is empty.
#
# The date 2012-10-17 is just the version of the policy language used here [1].
#
# [1]: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_version.html
resource "aws_iam_role" "goaswtf" {
  name               = "goaswtf"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }
}
POLICY
}