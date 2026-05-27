# Create a Bedrock AgentCore MCP gateway with CUSTOM_JWT authorization.
# The resource automatically resolves the chicken-and-egg problem: it creates
# the gateway with a placeholder audience, then immediately patches allowed_audience
# to equal the gateway's own gateway_id (required for Connect AppIntegrations).
resource "awsext_bedrockagentcore_gateway" "example" {
  name        = "bc-sonic-addservice-gateway"
  description = "AgentCore MCP gateway for AddService"
  role_arn    = "arn:aws:iam::123456789012:role/bedrock-agentcore-gateway-role"

  custom_jwt_authorizer = {
    discovery_url = "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_EXAMPLE/.well-known/openid-configuration"
  }

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}
