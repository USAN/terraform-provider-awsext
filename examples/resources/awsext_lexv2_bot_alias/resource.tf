resource "awsext_lexv2_bot_alias" "example" {
  bot_id      = awsext_lexv2_bot_import.example.bot_id
  name        = "prod"
  bot_version = "1"
  locale_id   = "en_US"

  text_log_enabled          = true
  text_log_cw_log_group_arn = "arn:aws:logs:us-east-1:123456789012:log-group:/aws/lex/bot-logs"
  text_log_prefix           = "lex/"

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}
