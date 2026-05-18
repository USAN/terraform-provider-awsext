# awsext_qconnect_ai_agent

Creates and manages an Amazon Q in Connect AI Agent (`qconnect:CreateAIAgent`). An AI Agent encapsulates the configuration that drives a specific behaviour (e.g. answer recommendation, manual search, self-service) within a Q in Connect assistant. Fields other than `configuration`, `visibility_status`, and `tags` force replacement when changed.

## Example Usage

```hcl
resource "awsext_qconnect_ai_agent" "customer_agent" {
  assistant_id      = awsext_qconnect_assistant.example.assistant_id
  name              = "customer-agent"
  type              = "ANSWER_RECOMMENDATION"
  visibility_status = "PUBLISHED"
  configuration = jsonencode({
    answerGenerationAIPromptId = awsext_qconnect_ai_prompt.example.ai_prompt_id
    locale                     = "en_US"
  })
}
```

## Argument Reference

The following arguments are supported:

- `assistant_id` - (Required, Forces new resource) Identifier of the parent Amazon Q in Connect assistant.
- `name` - (Required, Forces new resource) Name of the AI Agent.
- `type` - (Required, Forces new resource) Type of the AI Agent. One of `ANSWER_RECOMMENDATION`, `MANUAL_SEARCH`, `SELF_SERVICE`, `EMAIL_OVERVIEW`, `EMAIL_RESPONSE`, `EMAIL_GENERATIVE_ANSWER`, `NOTE_TAKING`, `ORCHESTRATION`, `CASE_SUMMARIZATION`.
- `configuration` - (Required) JSON string containing the `AIAgentConfiguration` payload for the chosen `type`. The JSON shape varies by agent type — refer to the [AWS Q in Connect API Reference](https://docs.aws.amazon.com/amazon-q-connect/latest/APIReference/) for the fields applicable to each type. Use `jsonencode({...})` in Terraform to avoid whitespace drift between plan and apply.
- `visibility_status` - (Optional) Visibility status of the AI Agent. One of `PUBLISHED` or `SAVED`. Defaults to `SAVED` when omitted.
- `description` - (Optional, Forces new resource) Description of the AI Agent.
- `tags` - (Optional) Tags to assign to the AI Agent.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `ai_agent_id` - Service-assigned UUID of the AI Agent.
- `ai_agent_arn` - ARN of the AI Agent.

## Import

AI Agents can be imported using the composite ID `<assistant_id>/<ai_agent_id>`:

```
terraform import awsext_qconnect_ai_agent.example <assistant_id>/<ai_agent_id>
```
