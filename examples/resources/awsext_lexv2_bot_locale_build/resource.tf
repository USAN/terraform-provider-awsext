resource "awsext_lexv2_bot_locale_build" "addservice_en_us" {
  bot_id      = awsext_lexv2_bot_import.addservice.bot_id
  bot_version = "DRAFT"
  locale_id   = "en_US"
  source_hash = "abc123"
}
