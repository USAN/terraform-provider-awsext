resource "awsext_qconnect_ai_prompt" "example" {
  assistant_id      = awsext_qconnect_assistant.example.assistant_id
  name              = "example-prompt"
  type              = "ANSWER_GENERATION"
  api_format        = "ANTHROPIC_CLAUDE_MESSAGES"
  model_id          = "anthropic.claude-3-5-sonnet-20240620-v1:0"
  template_type     = "TEXT"
  visibility_status = "PUBLISHED"

  template_configuration = <<-YAML
    system: You are a helpful assistant.
    messages:
      - role: user
        content: "{{question}}"
  YAML
}
