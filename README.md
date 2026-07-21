# Terraform Provider - AWS Extras

This is an AWS **Extras** provider. It provides various extras that the default/primary provider does not.

## Resources

### Connect

- `awsext_connect_agent_status` — manage Connect agent status values

### AppStream (WorkSpaces Applications)

- `awsext_appstream_image_copy` — copy an AppStream image from a source region into the provider's region (`CopyImage`). Owner-account only. Create waits for the destination image to reach `AVAILABLE` (default 30 min, configurable via `create_timeout_minutes`); delete retries on in-use errors with backoff (default 10 min, configurable via `delete_timeout_minutes`). Import by destination image name.
- `awsext_appstream_image_permission` — share a private AppStream image with another account (`UpdateImagePermissions` / `DeleteImagePermissions`); one instance per (image, account) pair with `allow_fleet` / `allow_image_builder` flags. Import as `<image_name>/<shared_account_id>`.

### WorkSpaces

- `awsext_workspaces_bundle` — create/update WorkSpaces bundles
- `awsext_workspaces_image` — create WorkSpaces images and manage image permissions
- `awsext_workspaces_image_copy` — copy a WorkSpaces image from a source region (`CopyWorkspaceImage`)
- `awsext_workspaces_image_permission` — share a WorkSpaces image with another account; import as `<image_id>/<shared_account_id>`
- `awsext_workspaces_pool` — manage WorkSpaces pools, including a `desired_state` (`STARTED`/`STOPPED`) attribute
- `awsext_workspaces_pool_running` — start/stop a WorkSpaces pool
- `awsext_workspaces_streaming_properties` — manage streaming properties on a WorkSpaces directory

> **Note:** AWS is retiring WorkSpaces Pools (EOL 2027-12-31); the AppStream resources above support the migration to WorkSpaces Applications.

## Example Usage

```terraform
resource "awsext_appstream_image_copy" "copy" {
  name              = "workbench-uw2" # destination image name (provider region is the destination)
  description       = "Workbench image copied to us-west-2"
  source_image_name = "workbench-test"
  source_region     = "us-east-1"
}

resource "awsext_appstream_image_permission" "share" {
  image_name          = "workbench-test"
  shared_account_id   = "111111111111"
  allow_fleet         = true
  allow_image_builder = false
}
```
