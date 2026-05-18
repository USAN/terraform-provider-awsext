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
