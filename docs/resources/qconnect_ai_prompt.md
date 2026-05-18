# awsext_qconnect_ai_prompt

Creates and manages an Amazon Q in Connect AI Prompt (`qconnect:CreateAIPrompt`). An AI Prompt defines the prompt template and model configuration used by a Q in Connect assistant for a specific purpose (e.g. answer generation, query reformulation). Fields other than `template_configuration`, `visibility_status`, and `tags` force replacement when changed.

## Example Usage

```hcl
resource "awsext_qconnect_ai_prompt" "example" {
  assistant_id  = awsext_qconnect_assistant.example.assistant_id
  name          = "example-prompt"
  type          = "ANSWER_GENERATION"
  api_format    = "ANTHROPIC_CLAUDE_MESSAGES"
  model_id      = "anthropic.claude-3-5-sonnet-20240620-v1:0"
  template_type = "TEXT"
  template_configuration = jsonencode({
    template = "You are an assistant. Respond to: {{question}}"
  })
  visibility_status = "PUBLISHED"
}
```

## Argument Reference

The following arguments are supported:

- `assistant_id` - (Required, Forces new resource) Identifier of the parent Amazon Q in Connect assistant.
- `name` - (Required, Forces new resource) Name of the AI Prompt. Must be between 1 and 255 characters.
- `type` - (Required, Forces new resource) Type of the AI Prompt, indicating its purpose. Examples: `ANSWER_GENERATION`, `QUERY_REFORMULATION`, `INTENT_LABELING_GENERATION`, `SELF_SERVICE_PRE_PROCESSING`, `SELF_SERVICE_ANSWER_GENERATION`, `ORCHESTRATION`, `NOTE_TAKING`.
- `api_format` - (Required, Forces new resource) API format for the AI Prompt. One of `ANTHROPIC_CLAUDE_MESSAGES`, `ANTHROPIC_CLAUDE_TEXT_COMPLETIONS`, `MESSAGES`, `TEXT_COMPLETIONS`. The `ANTHROPIC_CLAUDE_*` variants are deprecated; prefer `MESSAGES` or `TEXT_COMPLETIONS`.
- `model_id` - (Required, Forces new resource) Identifier of the model used for this AI Prompt (e.g. `anthropic.claude-3-5-sonnet-20240620-v1:0`).
- `template_type` - (Required, Forces new resource) Type of the prompt template. Currently the only supported value is `TEXT`.
- `template_configuration` - (Required) JSON document representing the template configuration. For `template_type = "TEXT"` this is a JSON object with a `text` key containing the prompt. Whitespace and key-order differences between the plan and the API-returned form are suppressed.
- `visibility_status` - (Optional) Visibility status of the AI Prompt. One of `PUBLISHED` or `SAVED`. Defaults to `SAVED` when omitted.
- `description` - (Optional, Forces new resource) Description of the AI Prompt.
- `tags` - (Optional) Tags to assign to the AI Prompt.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `ai_prompt_id` - Service-assigned UUID of the AI Prompt.
- `ai_prompt_arn` - ARN of the AI Prompt.

## Import

AI Prompts can be imported using the composite ID `<assistant_id>/<ai_prompt_id>`:

```
terraform import awsext_qconnect_ai_prompt.example <assistant_id>/<ai_prompt_id>
```
