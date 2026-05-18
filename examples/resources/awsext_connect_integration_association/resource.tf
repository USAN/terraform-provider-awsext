# Associate a Q Connect (Wisdom) Assistant with a Connect instance.
# This enables the AI-powered agent assistance feature in Connect.
resource "awsext_connect_integration_association" "wisdom" {
  instance_id      = "d9519e8f-2f9f-4a37-bf09-4bda8e27185d"
  integration_type = "WISDOM_ASSISTANT"
  integration_arn  = awsext_qconnect_assistant.example.assistant_arn
}

# Associate an AppIntegrations Application (e.g. an AgentCore MCP gateway)
# with a Connect instance. The official aws_connect_integration_association
# resource does not support the APPLICATION integration_type; use this resource instead.
resource "awsext_connect_integration_association" "mcp_app" {
  instance_id      = "d9519e8f-2f9f-4a37-bf09-4bda8e27185d"
  integration_type = "APPLICATION"
  integration_arn  = awsext_appintegrations_application.gateway.application_arn

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}
