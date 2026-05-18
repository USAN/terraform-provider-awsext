resource "awsext_qconnect_ai_guardrail_version" "v" {
  assistant_id    = awsext_qconnect_assistant.example.assistant_id
  ai_guardrail_id = awsext_qconnect_ai_guardrail.example.ai_guardrail_id
}
