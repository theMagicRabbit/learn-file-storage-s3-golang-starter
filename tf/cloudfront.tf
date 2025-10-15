resource "aws_cloudfront_distribution" "tubely_cloudfront" {
  comment = "cache for tubely"

  origin {
    domain_name              = aws_s3_bucket.private_bucket.bucket_regional_domain_name
    origin_id                = "tubelycdn"
    origin_access_control_id = aws_cloudfront_origin_access_control.tubely_oac.id
  }

  default_cache_behavior {
    allowed_methods        = ["HEAD", "GET"]
    cached_methods         = ["HEAD", "GET"]
    compress               = true
    target_origin_id       = "tubelycdn"
    smooth_streaming       = true
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = "658327ea-f89d-4fab-a63d-7e88639e58f6"
  }

  price_class = "PriceClass_All"
  enabled     = true

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
    ssl_support_method             = "vip"
    minimum_protocol_version       = "TLSv1"
  }

  is_ipv6_enabled = true
  http_version    = "http2"
  staging         = false
}

resource "aws_cloudfront_origin_access_control" "tubely_oac" {
  name                              = "tubely-oac"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
  origin_access_control_origin_type = "s3"
}

resource "aws_s3_bucket_policy" "tubely_cloudfront_s3_policy" {
  bucket = aws_s3_bucket.private_bucket.id
  policy = data.aws_iam_policy_document.tubely_cloudfront_bucket_access_document.json
}

data "aws_iam_policy_document" "tubely_cloudfront_bucket_access_document" {
  statement {
    sid    = "AllowCloudfrontAccessToPrivateBucket"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    resources = ["${aws_s3_bucket.private_bucket.arn}/*"]

    actions = ["s3:GetObject"]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.tubely_cloudfront.arn]
    }
  }
}


