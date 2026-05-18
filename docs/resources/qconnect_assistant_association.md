# awsext_qconnect_assistant_association

Creates and manages an Amazon Q in Connect assistant association. Wraps `qconnect:CreateAssistantAssociation` / `GetAssistantAssociation` / `DeleteAssistantAssociation`. No update API exists for structural fields; all are ForceNew except `tags`, which is updatable in place.

An association wires either a native Q in Connect knowledge base or an external Bedrock knowledge base to an assistant. The `association_type` attribute controls which sub-field of the `association` block is used.

## Example Usage

### External Bedrock Knowledge Base

```hcl
resource "awsext_qconnect_assistant_association" "bedrock" {
  assistant_id     = awsext_qconnect_assistant.example.assistant_id
  association_type = "EXTERNAL_BEDROCK_KNOWLEDGE_BASE"

  association = {
    bedrock_knowledge_base = {
      knowledge_base_arn = "arn:aws:bedrock:us-east-1::knowledge-base/ABCDEF1234"
      access_role_arn    = "arn:aws:iam::111111111111:role/QConnectKBRole"
    }
  }
}
```

### Native Q in Connect Knowledge Base

```hcl
resource "awsext_qconnect_assistant_association" "native_kb" {
  assistant_id     = awsext_qconnect_assistant.example.assistant_id
  association_type = "KNOWLEDGE_BASE"

  association = {
    knowledge_base_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  }
}
```

## Argument Reference

- `assistant_id` (Required, ForceNew) - Identifier of the Q in Connect assistant.
- `association_type` (Required, ForceNew) - Type of association. One of `KNOWLEDGE_BASE` or `EXTERNAL_BEDROCK_KNOWLEDGE_BASE`. Controls which sub-field of `association` is used.
- `association` (Required, ForceNew) - Block describing the associated resource. Exactly one sub-field should be set based on `association_type`:
  - `knowledge_base_id` (Optional) - Identifier of the Q in Connect knowledge base. Used when `association_type` is `KNOWLEDGE_BASE`.
  - `bedrock_knowledge_base` (Optional) - Configuration for an external Bedrock knowledge base. Used when `association_type` is `EXTERNAL_BEDROCK_KNOWLEDGE_BASE`.
    - `knowledge_base_arn` (Required) - ARN of the external Bedrock knowledge base.
    - `access_role_arn` (Required) - ARN of the IAM role used to access the external Bedrock knowledge base.
- `tags` (Optional) - Map of tags. Updatable in place without replacement.

## Attribute Reference

- `assistant_association_id` - Service-assigned UUID of the association.
- `assistant_association_arn` - ARN of the association.

## Import

Import using `<assistant_id>/<assistant_association_id>`:

```
terraform import awsext_qconnect_assistant_association.example <assistant_id>/<assistant_association_id>
```
