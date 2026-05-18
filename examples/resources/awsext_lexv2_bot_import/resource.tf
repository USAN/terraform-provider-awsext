resource "awsext_lexv2_bot_import" "addservice" {
  zip_file_path = "${path.module}/../_Resource/Connect/Lex/bot.zip"
  zip_sha256    = filesha256("${path.module}/../_Resource/Connect/Lex/bot.zip")
  role_arn      = "arn:aws:iam::123456789012:role/lex-bot-role"
  merge_strategy = "Overwrite"

  import_resource_specification = jsonencode({
    botImportSpecification = {
      botName                 = "bc-sonic-dev-AIAgent-Bot"
      idleSessionTTLInSeconds = 300
      roleArn                 = "arn:aws:iam::123456789012:role/lex-bot-role"
      dataPrivacy             = { childDirected = false }
    }
  })
}
