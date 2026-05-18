resource "awsext_connect_security_profile" "aiagent_bot" {
  instance_id = "d9519e8f-2f9f-4a37-bf09-4bda8e27185d"
  name        = "bc-sonic-addservice-bot"
  description = "Security profile for the add service bot"
  permissions = ["BasicAgentAccess", "OutboundCallAccess"]

  applications = [
    {
      namespace               = "CCP"
      application_permissions = ["ACCESS"]
    },
    {
      namespace               = "AgentApplications"
      application_permissions = ["ACCESS"]
    },
    {
      namespace = "bc-sonic-addservice-api-emyj-nmpnce5jah"
      application_permissions = [
        "aiagenttools___context_get",
        "aiagenttools___context_update",
        "aiagenttools___get_date",
        "sapapi___processmoveinrequest",
      ]
    },
  ]
}
