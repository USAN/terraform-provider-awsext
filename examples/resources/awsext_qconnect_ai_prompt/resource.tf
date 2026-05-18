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
