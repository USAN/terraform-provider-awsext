# Associate a Lex V2 bot alias with a Connect instance.
# Use this resource instead of aws_connect_bot_association, which only supports Lex V1.
resource "awsext_connect_bot_association" "example" {
  instance_id = "d9519e8f-2f9f-4a37-bf09-4bda8e27185d"
  alias_arn   = awsext_lexv2_bot_alias.example.arn
}
