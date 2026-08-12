resource "awsext_connect_security_profile_association" "aiagent_bot" {
  instance_id         = "d9519e8f-2f9f-4a37-bf09-4bda8e27185d"
  security_profile_id = awsext_connect_security_profile.aiagent_bot.id
  entity_type         = "AI_AGENT"
  entity_arn          = "arn:aws:connect:us-east-1:123456789012:instance/d9519e8f-2f9f-4a37-bf09-4bda8e27185d/ai-agent/abcd1234-5678-90ef-ghij-klmnopqrstuv"
}
