# awsext_connect_integration_association

Creates and manages an Amazon Connect Integration Association — a link between an Amazon Connect instance and an external AWS resource such as an Amazon Q Connect (Wisdom) Assistant or an AppIntegrations Application.

This resource fills the gap left by the official `aws_connect_integration_association` resource, which does not support the `APPLICATION` or `WISDOM_ASSISTANT` integration types.

**APIs used:**
- `connect:CreateIntegrationAssociation` — Create
- `connect:ListIntegrationAssociations` — Read (paginated, filtered by `integration_type`)
- `connect:DeleteIntegrationAssociation` — Delete
- `connect:TagResource` / `connect:UntagResource` / `connect:ListTagsForResource` — tag management

## Example Usage

### Wisdom (Q Connect) Assistant

```hcl
resource "awsext_connect_integration_association" "wisdom" {
  instance_id      = "d9519e8f-2f9f-4a37-bf09-4bda8e27185d"
  integration_type = "WISDOM_ASSISTANT"
  integration_arn  = awsext_qconnect_assistant.example.assistant_arn
}
```

### AppIntegrations Application (MCP gateway)

```hcl
resource "awsext_connect_integration_association" "mcp_app" {
  instance_id      = "d9519e8f-2f9f-4a37-bf09-4bda8e27185d"
  integration_type = "APPLICATION"
  integration_arn  = awsext_appintegrations_application.gateway.application_arn

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}
```

## Argument Reference

### Required

- `instance_id` — (Required, Forces replacement) The identifier of the Amazon Connect instance.
- `integration_type` — (Required, Forces replacement) The type of integration. Valid values: `EVENT`, `VOICE_ID`, `PINPOINT_APP`, `WISDOM_ASSISTANT`, `WISDOM_KNOWLEDGE_BASE`, `WISDOM_QUICK_RESPONSES`, `Q_MESSAGE_TEMPLATES`, `CASES_DOMAIN`, `APPLICATION`, `FILE_SCANNER`, `SES_IDENTITY`, `ANALYTICS_CONNECTOR`, `CALL_TRANSFER_CONNECTOR`, `COGNITO_USER_POOL`.
- `integration_arn` — (Required, Forces replacement) The ARN of the resource to associate.

### Optional

- `source_application_url` — (Optional, Forces replacement) The URL of the external application. Required when `integration_type` is `EVENT`.
- `source_application_name` — (Optional, Forces replacement) The name of the external application. Required when `integration_type` is `EVENT`.
- `source_type` — (Optional, Forces replacement) The data source type. Required when `integration_type` is `EVENT`. Valid values: `SALESFORCE`, `ZENDESK`, `CASES`.
- `tags` — (Optional) A map of tags to assign to the resource. Updatable in place.

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `integration_association_id` — The service-assigned identifier for the integration association.
- `integration_association_arn` — The ARN of the integration association.

## Import

Import an existing integration association using the format `<instance_id>/<integration_association_id>`:

```bash
terraform import awsext_connect_integration_association.example d9519e8f-2f9f-4a37-bf09-4bda8e27185d/12345678-1234-1234-1234-123456789012
```

After import, ensure your Terraform configuration includes the correct `integration_type` value; it is needed during the subsequent `terraform refresh` to filter the list API response correctly.
