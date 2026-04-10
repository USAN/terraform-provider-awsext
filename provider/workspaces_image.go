package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	workspacestypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &WorkspacesImageResource{}
var _ resource.ResourceWithImportState = &WorkspacesImageResource{}

func NewWorkspacesImageResource() resource.Resource {
	return &WorkspacesImageResource{}
}

type WorkspacesImageResource struct {
	config aws.Config
}

type WorkspacesImageResourceModel struct {
	ImageId        types.String `tfsdk:"image_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	WorkspaceId    types.String `tfsdk:"workspace_id"`
	State          types.String `tfsdk:"state"`
	OwnerAccountId types.String `tfsdk:"owner_account_id"`
}

func (r *WorkspacesImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspaces_image"
}

func (r *WorkspacesImageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a WorkSpace image from an existing stopped WorkSpace via CreateWorkspaceImage.",

		Attributes: map[string]schema.Attribute{
			"image_id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the new WorkSpace image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the new WorkSpace image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "The description of the new WorkSpace image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "The identifier of the source WorkSpace. The WorkSpace must be in the STOPPED state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The availability status of the image (AVAILABLE, PENDING, ERROR).",
			},
			"owner_account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the AWS account that owns the image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *WorkspacesImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(aws.Config)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *aws.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.config = config
}

func (r *WorkspacesImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkspacesImageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := workspaces.NewFromConfig(r.config)

	output, err := conn.CreateWorkspaceImage(ctx, &workspaces.CreateWorkspaceImageInput{
		Name:        aws.String(data.Name.ValueString()),
		Description: aws.String(data.Description.ValueString()),
		WorkspaceId: aws.String(data.WorkspaceId.ValueString()),
	})

	if err != nil {
		resp.Diagnostics.AddError("Error creating WorkSpace image",
			fmt.Sprintf("Could not create WorkSpace image, unexpected error: %s", err))
		return
	}

	tflog.Trace(ctx, "created workspaces image resource")

	data.ImageId = types.StringValue(aws.ToString(output.ImageId))
	data.State = types.StringValue(string(output.State))
	data.OwnerAccountId = types.StringValue(aws.ToString(output.OwnerAccountId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspacesImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkspacesImageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := workspaces.NewFromConfig(r.config)

	output, err := conn.DescribeWorkspaceImages(ctx, &workspaces.DescribeWorkspaceImagesInput{
		ImageIds: []string{data.ImageId.ValueString()},
	})

	if err != nil {
		var notFound *workspacestypes.ResourceNotFoundException
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading WorkSpace image",
			fmt.Sprintf("Could not read WorkSpace image %s, unexpected error: %s", data.ImageId.ValueString(), err))
		return
	}

	if len(output.Images) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	img := output.Images[0]
	data.ImageId = types.StringValue(aws.ToString(img.ImageId))
	data.Name = types.StringValue(aws.ToString(img.Name))
	data.Description = types.StringValue(aws.ToString(img.Description))
	data.State = types.StringValue(string(img.State))
	data.OwnerAccountId = types.StringValue(aws.ToString(img.OwnerAccountId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkspacesImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All mutable attributes use RequiresReplace; Update is never called.
}

func (r *WorkspacesImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkspacesImageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := workspaces.NewFromConfig(r.config)

	_, err := conn.DeleteWorkspaceImage(ctx, &workspaces.DeleteWorkspaceImageInput{
		ImageId: aws.String(data.ImageId.ValueString()),
	})

	if err != nil {
		var notFound *workspacestypes.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting WorkSpace image",
			fmt.Sprintf("Could not delete WorkSpace image %s, unexpected error: %s", data.ImageId.ValueString(), err))
		return
	}
}

func (r *WorkspacesImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("image_id"), req, resp)
}
