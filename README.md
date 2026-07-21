# Terraform Provider - AWS Extras

This is an AWS **Extras** provider (`registry.terraform.io/usan/awsext`). It exposes AWS API
operations missing from the official `hashicorp/aws` provider, so USAN does not need to
maintain a fork of `terraform-provider-aws`.

## Stack & architecture

- Go, module path `github.com/USAN/terraform-provider-awsext`
- Terraform Plugin Framework (not the legacy plugin SDK v2)
- AWS SDK for Go v2 (`workspaces`, `appstream`, `connect`, `sts` service packages)
- Provider `Configure` loads an `aws.Config` (static keys, named profile, or STS assume-role)
  and passes it as `ResourceData`; each resource receives it via `req.ProviderData.(aws.Config)`
- Custom retry policy: 20 max attempts, 10-second max backoff
- Strongly-typed AWS error handling via `errors.As`; resources remove themselves from state on
  not-found
- Shared WorkSpaces tag helpers live in `provider/workspaces_tags.go`
  (`readResourceTags`, `updateResourceTags`, `tagsToList`)
- No acceptance tests currently implemented

## Resources

### Connect

- `awsext_connect_agent_status` — manage Connect agent status values.
  APIs: `CreateAgentStatus`, `DescribeAgentStatus`, `UpdateAgentStatus`, `ListAgentStatuses`.
  Supports an `import_on_exists` write-only flag to adopt an existing status instead of
  erroring on create. Identity schema: `arn` + `agent_status_id`.

### AppStream (WorkSpaces Applications)

- `awsext_appstream_image_copy` — copy an AppStream image from a source region into the
  provider's region. Owner-account only (cannot copy an image merely shared to you).
  APIs: `CopyImage` (issued against the source region with the provider region as
  destination), `DescribeImages`, `DeleteImage`, `TagResource`/`UntagResource`/`ListTagsForResource`.
  Create waits for the destination image to reach `AVAILABLE` (default 30 min,
  `create_timeout_minutes`); delete retries in-use errors with backoff (default 10 min,
  `delete_timeout_minutes`). Import by destination image name.
- `awsext_appstream_image_permission` — share a private AppStream image with another account;
  one instance per (image, account) pair, keyed by image **name**.
  APIs: `UpdateImagePermissions` (create/update — upserts per account),
  `DeleteImagePermissions`, `DescribeImagePermissions` (paginated read).
  `allow_fleet` / `allow_image_builder` flags update in place.
  Import as `<image_name>/<shared_account_id>`.

### WorkSpaces

- `awsext_workspaces_bundle` — full CRUD via `CreateWorkspaceBundle`,
  `DescribeWorkspaceBundles`, `UpdateWorkspaceBundle`, `DeleteWorkspaceBundle`.
  `image_id` and tags updatable in place; other fields force replacement. Import by `bundle_id`.
- `awsext_workspaces_image` — `CreateWorkspaceImage`, `DescribeWorkspaceImages`,
  `DeleteWorkspaceImage`. All user-facing fields force replacement (no update API);
  tags updatable in place. Import by `image_id`.
- `awsext_workspaces_image_copy` — `CopyWorkspaceImage`, `DescribeWorkspaceImages`,
  `DeleteWorkspaceImage`. Describes the image immediately after copy to populate
  `owner_account_id`. Import by `image_id`.
- `awsext_workspaces_image_permission` — `UpdateWorkspaceImagePermission`
  (`AllowCopyImage=true` on create, `false` on delete), `DescribeWorkspaceImagePermissions`.
  Import as `<image_id>/<shared_account_id>`.
- `awsext_workspaces_pool` — full pool lifecycle via `CreateWorkspacesPool`,
  `DescribeWorkspacesPools`, `UpdateWorkspacesPool`, `TerminateWorkspacesPool` (permanent
  removal). Nested `application_settings` and `timeout_settings`. Import by `pool_id`.
- `awsext_workspaces_pool_running` — run-state only, via `StartWorkspacesPool` /
  `StopWorkspacesPool`; catches `InvalidResourceStateException`.
- `awsext_workspaces_streaming_properties` — `ModifyStreamingProperties` (create/update),
  read via `DescribeWorkspaceDirectories`; delete is a no-op (no AWS delete API).
  Import by `directory_id`.

> **Note:** AWS is retiring WorkSpaces Pools (EOL 2027-12-31); the AppStream resources above
> support the migration to WorkSpaces Applications.

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

## Local development

Build locally and point Terraform at the build with a `dev_overrides` block in your CLI
config (`%APPDATA%\terraform.rc` on Windows):

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/usan/awsext" = "C:\\Users\\<you>\\go\\bin"
  }
  direct {}
}
```

Then `go install .` and run `terraform plan`/`apply` directly (skip `terraform init` for the
overridden provider).

## Conventions for new resources

- `var _ resource.Resource = &XxxResource{}` interface assertions at the top of each file;
  add `resource.ResourceWithImportState` when import is supported
- Resource struct holds `config aws.Config`; `Configure` asserts
  `req.ProviderData.(aws.Config)` and returns early on nil
- Service clients created per call: `workspaces.NewFromConfig(r.config)` /
  `appstream.NewFromConfig(r.config)`
- Not-found errors remove the resource from state (`resp.State.RemoveResource`)
- Register new resources in `provider.go` `Resources()`
