# Terraform Provider - AWS Extras

A Terraform provider (`registry.terraform.io/usan/awsext`) that fills gaps in the official `hashicorp/aws` provider.

## Purpose

Some AWS API operations have no resource in the official provider.
This provider exposes those operations as first-class Terraform resources, so USAN does not need to maintain a fork of `terraform-provider-aws`.

## Features

### Connect

- `awsext_connect_agent_status` — manages Connect agent status values.
  APIs: `CreateAgentStatus`, `DescribeAgentStatus`, `UpdateAgentStatus`, and `ListAgentStatuses`.
  Supports an `import_on_exists` write-only flag to adopt an existing status instead of erroring on create.
  Identity schema: `arn` + `agent_status_id`.

### AppStream (WorkSpaces Applications)

- `awsext_appstream_image_copy` — copies an AppStream image from a source region into the provider's region.
  Owner-account only; an image merely shared to the account cannot be copied.
  APIs: `CopyImage` (issued against the source region with the provider region as destination), `DescribeImages`, `DeleteImage`, and `TagResource`/`UntagResource`/`ListTagsForResource`.
  Create waits for the destination image to reach `AVAILABLE` (default 30 minutes, `create_timeout_minutes`).
  Delete retries in-use errors with backoff (default 10 minutes, `delete_timeout_minutes`).
  Import by destination image name.
- `awsext_appstream_image_permission` — shares a private AppStream image with another account.
  One instance manages one (image, account) pair, keyed by image **name**.
  APIs: `UpdateImagePermissions` (create/update — upserts per account), `DeleteImagePermissions`, and `DescribeImagePermissions` (paginated read).
  The `allow_fleet` and `allow_image_builder` flags update in place.
  Import as `<image_name>/<shared_account_id>`.

### WorkSpaces

- `awsext_workspaces_bundle` — full CRUD via `CreateWorkspaceBundle`, `DescribeWorkspaceBundles`, `UpdateWorkspaceBundle`, and `DeleteWorkspaceBundle`.
  `image_id` and tags update in place; other fields force replacement.
  Import by `bundle_id`.
- `awsext_workspaces_image` — `CreateWorkspaceImage`, `DescribeWorkspaceImages`, and `DeleteWorkspaceImage`.
  All user-facing fields force replacement (no update API); tags update in place.
  Import by `image_id`.
- `awsext_workspaces_image_copy` — `CopyWorkspaceImage`, `DescribeWorkspaceImages`, and `DeleteWorkspaceImage`.
  Describes the image immediately after copy to populate `owner_account_id`.
  Import by `image_id`.
- `awsext_workspaces_image_permission` — `UpdateWorkspaceImagePermission` (`AllowCopyImage=true` on create, `false` on delete) and `DescribeWorkspaceImagePermissions`.
  Import as `<image_id>/<shared_account_id>`.
- `awsext_workspaces_pool` — full pool lifecycle via `CreateWorkspacesPool`, `DescribeWorkspacesPools`, `UpdateWorkspacesPool`, and `TerminateWorkspacesPool` (permanent removal).
  Nested `application_settings` and `timeout_settings`.
  Import by `pool_id`.
- `awsext_workspaces_pool_running` — run-state only, via `StartWorkspacesPool` and `StopWorkspacesPool`; catches `InvalidResourceStateException`.
- `awsext_workspaces_streaming_properties` — `ModifyStreamingProperties` (create/update); read via `DescribeWorkspaceDirectories`; delete is a no-op (no AWS delete API).
  Import by `directory_id`.

> **Note:** AWS is retiring WorkSpaces Pools (EOL 2027-12-31).
> The AppStream resources above support the migration to WorkSpaces Applications.

## Usage

```terraform
terraform {
  required_providers {
    awsext = {
      source  = "usan/awsext"
      version = ">= 1.4.0"
    }
  }
}

provider "awsext" {
  region  = "us-west-2"
  profile = "my-profile"
}

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

Import an existing share:

```console
terraform import awsext_appstream_image_permission.share workbench-test/111111111111
```

## Configuration

The provider accepts standard AWS credential settings; all are optional and fall back to the AWS default credential chain.

| Attribute    | Description                                                       |
|--------------|-------------------------------------------------------------------|
| `region`     | AWS region (destination region for `awsext_appstream_image_copy`) |
| `profile`    | Named AWS profile                                                 |
| `access_key` | Static access key (pair with `secret_key`)                        |
| `secret_key` | Static secret key                                                 |
| `token`      | Session token for temporary credentials                           |
| `role_arn`   | IAM role to assume via STS                                        |

Never commit static credentials; prefer SSO profiles or assumed roles.

## Dependencies

- Go (version per `go.mod`) to build
- Terraform Plugin Framework (not the legacy plugin SDK v2)
- AWS SDK for Go v2 (`appstream`, `connect`, `workspaces`, and `sts` service packages)

## Development

Build, test, and lint locally:

```console
go build ./...
go test ./...
golangci-lint run ./...
```

The lint config follows the USAN baseline (golangci-lint v1 schema); install golangci-lint 1.64.x:

```console
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

To run a local build against real Terraform configs, add a `dev_overrides` block to your CLI config (`%APPDATA%\terraform.rc` on Windows):

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/usan/awsext" = "C:\\Users\\<you>\\go\\bin"
  }
  direct {}
}
```

Then run `go install .` and use `terraform plan`/`apply` directly (skip `terraform init` for the overridden provider).

Releases are cut by pushing a `v*` tag; GitHub Actions runs GoReleaser with GPG signing.

Conventions for new resources:

- `var _ resource.Resource = &XxxResource{}` interface assertions at the top of each file; add `resource.ResourceWithImportState` when import is supported.
- The resource struct holds `config aws.Config`; `Configure` asserts `req.ProviderData.(aws.Config)` and returns early on nil.
- Service clients are created per call: `appstream.NewFromConfig(r.config)`.
- Not-found errors remove the resource from state (`resp.State.RemoveResource`).
- Register new resources in `provider.go` `Resources()`.
- Unit-test pure logic against narrow, consumer-side client interfaces (see `appstream_image_copy_test.go`).

## Attribution

This project was generated with AI assistance.

**Tool:** Claude (claude-fable-5)
**Skills:** usan-code v1.1.0.719339
**Date:** 2026-07-21
