# awsext_appintegrations_application

Creates and manages an Amazon AppIntegrations Application. Wraps `appintegrations:CreateApplication` / `GetApplication` / `UpdateApplication` / `DeleteApplication`. Used to register AgentCore gateway MCP servers as Amazon Connect workspace applications.

## Example Usage

```hcl
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
```

## Argument Reference

### Required

- `name` (String) - The name of the application. Updatable in place.
- `namespace` (String, ForceNew) - The namespace of the application. Must match the AgentCore gateway ID. Forces replacement on change.
- `application_source_config` (Object, ForceNew) - Configuration for where the application is loaded from. Forces replacement. Contains:
  - `external_url_config` (Object, Required) - External URL source configuration:
    - `access_url` (String, Required) - The URL used to access the application.
    - `approved_origins` (List of String, Optional) - Additional approved origin URLs.

### Optional

- `description` (String) - Description of the application. Updatable in place.
- `application_type` (String, ForceNew) - Type of application. One of `STANDARD`, `SERVICE`, `MCP_SERVER`. Forces replacement on change. Use `MCP_SERVER` for AgentCore gateway integrations.
- `permissions` (List of String) - Events or requests the application has access to. Updatable in place.
- `publications` (List of Object, ForceNew) - Events the application publishes. Deprecated by AWS in favor of `permissions`. Forces replacement. Each object contains:
  - `event` (String, Required) - Publication event name.
  - `schema` (String, Required) - JSON schema of the publication event.
  - `description` (String, Optional) - Description of the publication.
- `subscriptions` (List of Object, ForceNew) - Events the application subscribes to. Deprecated by AWS in favor of `permissions`. Forces replacement. Each object contains:
  - `event` (String, Required) - Subscription event name.
  - `description` (String, Optional) - Description of the subscription.
- `initialization_timeout` (Number) - Maximum time in milliseconds to establish a workspace connection. Updatable in place.
- `iframe_config` (Object) - Iframe configuration. Updatable in place. Contains:
  - `allow` (List of String, Optional) - Features allowed in the iframe.
  - `sandbox` (List of String, Optional) - Sandbox attributes for the iframe.
- `tags` (Map of String) - Tags to assign to the application.

## Attribute Reference

- `application_id` - Service-assigned unique identifier for the Application.
- `application_arn` - Amazon Resource Name (ARN) of the Application.

## Import

Import by Application ARN:

```
terraform import awsext_appintegrations_application.example arn:aws:app-integrations:us-east-1:123456789012:application/<application-id>
```
