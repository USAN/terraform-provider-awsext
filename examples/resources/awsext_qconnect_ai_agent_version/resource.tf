resource "awsext_qconnect_ai_agent_version" "v" {
  assistant_id = awsext_qconnect_assistant.example.assistant_id
  ai_agent_id  = awsext_qconnect_ai_agent.example.ai_agent_id
}
