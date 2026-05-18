# awsext_connect_security_profile

Creates and manages an Amazon Connect Security Profile.

This resource fills the gap left by the official `aws_connect_security_profile` resource, which does not expose the `applications` field. The `applications` field is required for Amazon Q Connect MCP (Model Context Protocol) wiring, where third-party application namespaces and their permissions must be granted to agents via their assigned security profile.

## Example Usage

```hcl
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
```

## Argument Reference

### Required

- `instance_id` (String) — The identifier of the Amazon Connect instance. Changing this forces a new resource.
- `name` (String) — The name of the security profile. Changing this forces a new resource.

### Optional

- `description` (String) — A description of the security profile.
- `permissions` (List of String) — The list of permissions granted to agents assigned to this security profile (e.g., `"BasicAgentAccess"`, `"OutboundCallAccess"`).
- `applications` (List of Object) — Third-party application namespaces and their permissions granted to agents. Each object contains:
  - `namespace` (String, Required) — The namespace of the third-party application.
  - `application_permissions` (List of String, Required) — The permissions granted within that namespace.
- `allowed_access_control_tags` (Map of String) — Tags that this security profile uses to restrict access to resources in Amazon Connect.
- `tag_restricted_resources` (List of String) — Resources that this security profile applies tag restrictions to in Amazon Connect.
- `tags` (Map of String) — AWS resource tags to assign to the security profile.

## Attribute Reference

- `security_profile_id` (String) — The service-assigned identifier for the security profile.
- `security_profile_arn` (String) — The Amazon Resource Name (ARN) for the security profile.

## Import

Import using `<instance_id>/<security_profile_id>`:

```shell
terraform import awsext_connect_security_profile.example d9519e8f-2f9f-4a37-bf09-4bda8e27185d/12345678-1234-1234-1234-123456789012
```
