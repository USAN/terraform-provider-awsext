# awsext_qconnect_ai_guardrail

Creates and manages an Amazon Q in Connect AI Guardrail (`qconnect:CreateAIGuardrail`). Guardrails define policies that block or filter harmful content, restricted topics, sensitive information, and custom words. The `name` and `description` attributes force replacement when changed. All policy configurations and messaging attributes are updatable in-place.

## Example Usage

```hcl
resource "awsext_qconnect_ai_guardrail" "example" {
  assistant_id              = awsext_qconnect_assistant.example.assistant_id
  name                      = "example-guardrail"
  blocked_input_messaging   = "I can't help with that."
  blocked_outputs_messaging = "I can't share that."

  content_policy_config = {
    filters_config = [
      { input_strength = "HIGH", output_strength = "HIGH", type = "SEXUAL" },
      { input_strength = "HIGH", output_strength = "HIGH", type = "VIOLENCE" },
    ]
  }
}
```

## Argument Reference

### Top-level Arguments

- `assistant_id` (Required, Forces new resource) — Identifier of the parent Amazon Q in Connect assistant.
- `name` (Required, Forces new resource) — Name of the AI Guardrail.
- `blocked_input_messaging` (Required) — Message returned to the user when the guardrail blocks an input prompt.
- `blocked_outputs_messaging` (Required) — Message returned to the user when the guardrail blocks a model response.
- `visibility_status` (Optional, Computed) — Visibility status. One of `PUBLISHED` or `SAVED`. Defaults to `SAVED`.
- `description` (Optional, Forces new resource) — Description of the AI Guardrail.
- `tags` (Optional, Computed) — Tags to assign to the resource.

### `content_policy_config` Block (Optional)

Contains harmful content filter policies.

- `filters_config` (Required) — List of content filter objects, each with:
  - `input_strength` (Required) — Filter strength for input prompts. One of `NONE`, `LOW`, `MEDIUM`, `HIGH`.
  - `output_strength` (Required) — Filter strength for model responses. One of `NONE`, `LOW`, `MEDIUM`, `HIGH`.
  - `type` (Required) — Content category. Examples: `SEXUAL`, `VIOLENCE`, `HATE`, `INSULTS`.

### `topic_policy_config` Block (Optional)

Defines topics that the guardrail should deny.

- `topics_config` (Required) — List of topic objects, each with:
  - `name` (Required) — Name of the topic to deny.
  - `definition` (Required) — Description of the topic that the guardrail should deny.
  - `type` (Required) — Topic type. Currently only `DENY` is supported.
  - `examples` (Optional) — List of example prompts that belong to this topic.

### `word_policy_config` Block (Optional)

Configures words and managed word lists to block.

- `words_config` (Optional) — List of custom word objects, each with:
  - `text` (Required) — Word text to block.
- `managed_word_lists_config` (Optional) — List of managed word list objects, each with:
  - `type` (Required) — Managed word list type. Currently only `PROFANITY` is supported.

### `sensitive_information_policy_config` Block (Optional)

Configures PII entity and regex-based sensitive information filtering.

- `pii_entities_config` (Optional) — List of PII entity configurations, each with:
  - `type` (Required) — PII entity type (e.g. `NAME`, `EMAIL`, `US_SOCIAL_SECURITY_NUMBER`).
  - `action` (Required) — Action when PII is detected. One of `BLOCK` or `ANONYMIZE`.
- `regexes_config` (Optional) — List of regex filter configurations, each with:
  - `name` (Required) — Name of the regex filter.
  - `pattern` (Required) — Regular expression pattern to match.
  - `action` (Required) — Action when the regex matches. One of `BLOCK` or `ANONYMIZE`.
  - `description` (Optional) — Description of the regex filter.

### `contextual_grounding_policy_config` Block (Optional)

Configures contextual grounding filters that block low-confidence responses.

- `filters_config` (Required) — List of grounding filter objects, each with:
  - `threshold` (Required) — Score threshold (0.0–1.0). Responses scoring below this value are blocked.
  - `type` (Required) — Filter type. One of `GROUNDING` or `RELEVANCE`.

## Attribute Reference

- `ai_guardrail_id` — Service-assigned UUID of the AI Guardrail.
- `ai_guardrail_arn` — ARN of the AI Guardrail.

## Import

Import using `<assistant_id>/<ai_guardrail_id>`:

```shell
terraform import awsext_qconnect_ai_guardrail.example <assistant_id>/<ai_guardrail_id>
```
