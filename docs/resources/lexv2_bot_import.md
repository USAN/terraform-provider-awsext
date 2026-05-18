# awsext_lexv2_bot_import

Imports an Amazon Lex V2 bot from a zip archive using the LexV2 import API
(`CreateUploadUrl` / `StartImport` / `DescribeImport`).

This resource fills the gap left by the official AWS provider, which does not support
importing a fully-configured bot definition from a zip archive.

## Create flow

1. If `import_on_exists` is `true`, the resource first calls `ListBots` to check whether
   a bot matching the name in `import_resource_specification` already exists. If found, the
   existing bot is adopted into Terraform state without running an import.
2. Calls `CreateUploadUrl` to obtain a pre-signed S3 URL and an `import_id`.
3. Reads the zip bytes from `zip_file_path` or decodes `zip_content_base64`.
4. HTTP-PUTs the zip to the pre-signed URL.
5. Calls `StartImport` with the provided `import_resource_specification` and `merge_strategy`.
6. Polls `DescribeImport` (5 s interval, up to 5 minutes) until the import status is
   `Completed` or `Failed`.
7. On success, calls `sts:GetCallerIdentity` to build the `bot_arn`.

## Example Usage

```hcl
resource "awsext_lexv2_bot_import" "addservice" {
  zip_file_path  = "${path.module}/../_Resource/Connect/Lex/bot.zip"
  zip_sha256     = filesha256("${path.module}/../_Resource/Connect/Lex/bot.zip")
  role_arn       = "arn:aws:iam::123456789012:role/lex-bot-role"
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
```

## Argument Reference

### Required

- `role_arn` (String) — IAM role ARN that Amazon Lex assumes to build and run the bot. Forces replacement.
- `merge_strategy` (String) — Determines how conflicts are resolved when the import already exists. Valid values: `Overwrite`, `FailOnConflict`, `Append`. Forces replacement.
- `import_resource_specification` (String) — JSON-encoded `ImportResourceSpecification` passed to `StartImport`. The JSON is normalized (sorted keys) to prevent cosmetic drift. Forces replacement.

### Optional

- `zip_file_path` (String, WriteOnly) — Local path to the zip archive containing the bot definition. Exactly one of `zip_file_path` or `zip_content_base64` must be set. Forces replacement.
- `zip_content_base64` (String, WriteOnly) — Base64-encoded zip archive. Exactly one of `zip_file_path` or `zip_content_base64` must be set. Forces replacement.
- `zip_sha256` (String) — SHA256 checksum of the zip archive. Stored in state; changing this value triggers replacement.
- `import_on_exists` (Bool, WriteOnly) — When `true`, if a bot matching the name in `import_resource_specification` already exists, adopt it into state without running a new import.

### Computed

- `bot_id` (String) — The service-assigned identifier of the imported bot.
- `bot_arn` (String) — The Amazon Resource Name (ARN) of the imported bot.
- `import_id` (String) — The import job identifier returned by `CreateUploadUrl`.
- `import_status` (String) — The final import status returned by `DescribeImport`.

## Import

Import an existing bot by its bot ID:

```shell
terraform import awsext_lexv2_bot_import.addservice <bot_id>
```

Note: `zip_file_path` / `zip_content_base64` and `import_on_exists` are write-only and cannot
be restored from import. You must add them back to your configuration manually.
