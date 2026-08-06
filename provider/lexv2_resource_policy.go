package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	lextypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &LexV2ResourcePolicyResource{}
var _ resource.ResourceWithImportState = &LexV2ResourcePolicyResource{}

func NewLexV2ResourcePolicyResource() resource.Resource {
	return &LexV2ResourcePolicyResource{}
}

type LexV2ResourcePolicyResource struct {
	config aws.Config
}

type LexV2ResourcePolicyModel struct {
	ResourceArn types.String `tfsdk:"resource_arn"`
	Policy      types.String `tfsdk:"policy"`
	RevisionId  types.String `tfsdk:"revision_id"`
}

func (r *LexV2ResourcePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lexv2_resource_policy"
}

func (r *LexV2ResourcePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a resource-based IAM policy to an Amazon Lex V2 bot or bot alias " +
			"(lexmodelsv2:CreateResourcePolicy / UpdateResourcePolicy / DeleteResourcePolicy).",

		Attributes: map[string]schema.Attribute{
			"resource_arn": schema.StringAttribute{
				Required:    true,
				Description: "ARN of the Lex V2 bot or bot alias. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.StringAttribute{
				Required:    true,
				Description: "JSON policy document to attach to the resource.",
			},
			"revision_id": schema.StringAttribute{
				Computed:    true,
				Description: "Current revision ID of the resource policy.",
			},
		},
	}
}

func (r *LexV2ResourcePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected aws.Config, got: %T.", req.ProviderData),
		)
		return
	}
	r.config = cfg
}

func (r *LexV2ResourcePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LexV2ResourcePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	createOut, err := client.CreateResourcePolicy(ctx, &lexmodelsv2.CreateResourcePolicyInput{
		ResourceArn: aws.String(data.ResourceArn.ValueString()),
		Policy:      aws.String(data.Policy.ValueString()),
	})
	if err != nil {
		var pre *lextypes.PreconditionFailedException
		if !errors.As(err, &pre) {
			resp.Diagnostics.AddError("Error creating Lex V2 resource policy", err.Error())
			return
		}
		// Policy already exists — adopt it by updating.
		updateOut, uerr := client.UpdateResourcePolicy(ctx, &lexmodelsv2.UpdateResourcePolicyInput{
			ResourceArn: aws.String(data.ResourceArn.ValueString()),
			Policy:      aws.String(data.Policy.ValueString()),
		})
		if uerr != nil {
			resp.Diagnostics.AddError("Error adopting existing Lex V2 resource policy", uerr.Error())
			return
		}
		data.RevisionId = types.StringPointerValue(updateOut.RevisionId)
		tflog.Trace(ctx, "adopted existing Lex V2 resource policy", map[string]any{"resource_arn": data.ResourceArn.ValueString()})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	data.RevisionId = types.StringPointerValue(createOut.RevisionId)
	tflog.Trace(ctx, "created Lex V2 resource policy", map[string]any{"resource_arn": data.ResourceArn.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LexV2ResourcePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LexV2ResourcePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	out, err := client.DescribeResourcePolicy(ctx, &lexmodelsv2.DescribeResourcePolicyInput{
		ResourceArn: aws.String(data.ResourceArn.ValueString()),
	})
	if err != nil {
		var nfe *lextypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			tflog.Warn(ctx, "Lex V2 resource policy not found, removing from state",
				map[string]any{"resource_arn": data.ResourceArn.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Lex V2 resource policy", err.Error())
		return
	}

	data.Policy = types.StringPointerValue(out.Policy)
	data.RevisionId = types.StringPointerValue(out.RevisionId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LexV2ResourcePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LexV2ResourcePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LexV2ResourcePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	out, err := client.UpdateResourcePolicy(ctx, &lexmodelsv2.UpdateResourcePolicyInput{
		ResourceArn:        aws.String(data.ResourceArn.ValueString()),
		Policy:             aws.String(data.Policy.ValueString()),
		ExpectedRevisionId: aws.String(state.RevisionId.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating Lex V2 resource policy", err.Error())
		return
	}

	data.RevisionId = types.StringPointerValue(out.RevisionId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LexV2ResourcePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LexV2ResourcePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	_, err := client.DeleteResourcePolicy(ctx, &lexmodelsv2.DeleteResourcePolicyInput{
		ResourceArn: aws.String(data.ResourceArn.ValueString()),
	})
	if err != nil {
		var nfe *lextypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("Error deleting Lex V2 resource policy", err.Error())
	}
}

func (r *LexV2ResourcePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_arn"), req.ID)...)
}
