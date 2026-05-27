# Attach a resource-based policy to a Lex V2 bot alias, allowing
# Amazon Connect to call RecognizeText on the alias.
resource "awsext_lexv2_resource_policy" "example" {
  resource_arn = awsext_lexv2_bot_alias.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = "connect.amazonaws.com" }
        Action    = "lex:RecognizeText"
        Resource  = awsext_lexv2_bot_alias.example.arn
        Condition = {
          StringEquals = {
            "AWS:SourceAccount" = "123456789012"
          }
        }
      }
    ]
  })
}
