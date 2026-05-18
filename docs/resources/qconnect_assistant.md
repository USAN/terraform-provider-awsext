# awsext_qconnect_assistant

Creates and manages an Amazon Q in Connect assistant. Wraps `qconnect:CreateAssistant` / `GetAssistant` / `DeleteAssistant`. No update API exists for non-tag fields; tags are updatable in place.

## Example Usage

```hcl
resource "awsext_qconnect_assistant" "example" {
  name        = "example-assistant"
  type        = "AGENT"
  description = "Example Q in Connect assistant"
  tags        = { env = "dev" }
}
```

## Argument Reference

- `name` (Required, ForceNew) - Friendly name of the assistant.
- `type` (Required, ForceNew) - Assistant type. Currently `AGENT`.
- `description` (Optional, ForceNew) - Description.
- `tags` (Optional) - Map of tags.

## Attribute Reference

- `assistant_id` - Service-assigned UUID.
- `assistant_arn` - ARN of the assistant.

## Import

```
terraform import awsext_qconnect_assistant.example <assistant_id>
```
