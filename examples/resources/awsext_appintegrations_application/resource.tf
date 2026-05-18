resource "awsext_appintegrations_application" "gateway" {
  name             = "bc-sonic-addservice-gateway"
  namespace        = "bc-sonic-addservice-api-emyj-nmpnce5jah"
  application_type = "MCP_SERVER"

  application_source_config = {
    external_url_config = {
      access_url = "https://bc-sonic-addservice-api-emyj-nmpnce5jah.gateway.bedrock-agentcore.us-east-1.amazonaws.com/mcp"
    }
  }

  tags = {
    env = "production"
  }
}
