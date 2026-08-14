resource "awsext_qconnect_ai_prompt_version" "v" {
  assistant_id = awsext_qconnect_assistant.example.assistant_id
  ai_prompt_id = awsext_qconnect_ai_prompt.example.ai_prompt_id
}
