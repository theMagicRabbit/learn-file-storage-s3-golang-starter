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

resource "aws_iam_group" "managers_group" {
  name = "managers"
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

resource "aws_iam_policy" "managers_from_home_policy" {
  name        = "manager-from-home"
  path        = "/"
  description = "Allows manager group access from home IP"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "VisualEditor0"
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
        Condition = {
          IpAddress = {
            "aws:SourceIp" = "99.137.90.132/32"
          }
        }
      }
    ]
  })
}

resource "aws_iam_policy" "tubely_s3_access_policy" {
  name = "tubely-s3"
  path = "/"
  description = "Access for tubely app"

  policy = jsonencode({
   Version =  "2012-10-17",
   Statement = [
     {
       Sid = "VisualEditor0",
       Effect = "Allow",
       Action = [
         "s3:PutObject",
         "s3:GetObject",
         "s3:DeleteObject",
         "s3:ListBucket"
       ],
       Resource = [aws_s3_bucket.learning_bucket.arn, "${aws_s3_bucket.learning_bucket.arn}/*"]
     }
    ]
  })
}

resource "aws_iam_group_policy_attachment" "attach_manager_from_home_to_manager" {
  group      = aws_iam_group.managers_group.name
  policy_arn = aws_iam_policy.managers_from_home_policy.arn
}

resource "aws_iam_group_membership" "managers_members" {
  name = "tubely-managers-group-memberships"

  users = ["tofu", "brt"]

  group = aws_iam_group.managers_group.name
}

