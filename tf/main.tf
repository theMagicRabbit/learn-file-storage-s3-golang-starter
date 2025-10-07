provider "aws" {
  region = "us-east-2"
}

resource "aws_s3_bucket" "learning_bucket" {
  bucket              = "tubely-22918"
  force_destroy       = true
  object_lock_enabled = false
}

resource "aws_s3_bucket_public_access_block" "allow_all_access" {
  bucket = aws_s3_bucket.learning_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "allow_get_access" {
  bucket = aws_s3_bucket.learning_bucket.id
  policy = <<EOP
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "${aws_s3_bucket.learning_bucket.arn}/*"
    }
  ]
}
EOP
}
