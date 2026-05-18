# awsext_qconnect_ai_guardrail_version

Creates a version snapshot of an Amazon Q in Connect AI Guardrail (`qconnect:CreateAIGuardrailVersion`). Versions are immutable point-in-time captures of the guardrail configuration. All inputs are ForceNew; any change causes Terraform to destroy and recreate the resource. No update API exists.

Use `guardrail_arn_with_version` in AI agent configurations that require a pinned guardrail version.

## Example Usage

```hcl
resource "awsext_qconnect_ai_guardrail_version" "v" {
  assistant_id    = awsext_qconnect_assistant.example.assistant_id
  ai_guardrail_id = awsext_qconnect_ai_guardrail.example.ai_guardrail_id
}
```

## Argument Reference

The following arguments are supported:

- `assistant_id` - (Required, ForceNew) Identifier of the parent Amazon Q in Connect assistant.
- `ai_guardrail_id` - (Required, ForceNew) Identifier of the AI Guardrail to version.
- `modified_time_seconds` - (Optional, ForceNew) Unix epoch seconds of the last-known modification time of the guardrail. When set, the API rejects the version create if the guardrail has been modified more recently, preventing accidental version creation on a stale configuration.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `version_number` - Service-assigned version number of the AI Guardrail version.
- `guardrail_arn_with_version` - ARN of the AI Guardrail qualified with the version number (`<ai_guardrail_arn>:<version_number>`). Use this in AI agent configurations that require a specific guardrail version.

## Import

AI Guardrail versions can be imported using the format `<assistant_id>/<ai_guardrail_id>/<version_number>`:

```
terraform import awsext_qconnect_ai_guardrail_version.v a1b2c3d4-5678-90ab-cdef-EXAMPLE11111/g1h2i3j4-5678-90ab-cdef-EXAMPLE22222/1
```
