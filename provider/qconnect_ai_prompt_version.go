package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qconnect"
	qconnecttypes "github.com/aws/aws-sdk-go-v2/service/qconnect/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &QConnectAIPromptVersionResource{}
var _ resource.ResourceWithImportState = &QConnectAIPromptVersionResource{}

func NewQConnectAIPromptVersionResource() resource.Resource {
	return &QConnectAIPromptVersionResource{}
}

type QConnectAIPromptVersionResource struct{ config aws.Config }

type QConnectAIPromptVersionResourceModel struct {
	AssistantId           types.String `tfsdk:"assistant_id"`
	AiPromptId            types.String `tfsdk:"ai_prompt_id"`
	ModifiedTimeSeconds   types.Int64  `tfsdk:"modified_time_seconds"`
	VersionNumber         types.Int64  `tfsdk:"version_number"`
	PromptArnWithVersion  types.String `tfsdk:"prompt_arn_with_version"`
	AiPromptIdWithVersion types.String `tfsdk:"ai_prompt_id_with_version"`
}

// Metadata

func (r *QConnectAIPromptVersionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qconnect_ai_prompt_version"
}

// Schema

func (r *QConnectAIPromptVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a version snapshot of an Amazon Q in Connect AI Prompt (qconnect:CreateAiPromptVersion). Versions are immutable; all inputs are ForceNew. No update API exists.",
		Attributes: map[string]schema.Attribute{
			"assistant_id": schema.StringAttribute{
				Required:      true,
				Description:   "Identifier of the parent Amazon Q in Connect Assistant.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ai_prompt_id": schema.StringAttribute{
				Required:      true,
				Description:   "Identifier of the AI prompt to version.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"modified_time_seconds": schema.Int64Attribute{
				Optional:      true,
				Description:   "Unix epoch seconds of the last-known modification time of the prompt. When set, the API rejects the version create if the prompt has been modified more recently, preventing accidental version creation on stale configs.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"version_number": schema.Int64Attribute{
				Computed:      true,
				Description:   "Service-assigned version number of the AI Prompt version.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"prompt_arn_with_version": schema.StringAttribute{
				Computed:      true,
				Description:   "Full ARN of the AI Prompt qualified with the version number (<ai_prompt_arn>:<version_number>)",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ai_prompt_id_with_version": schema.StringAttribute{
				Computed:      true,
				Description:   "AI Prompt ID qualified with the version number (<ai_prompt_id>:<version_number>). Use this for orchestrationPromptId in ORCHESTRATION AI agent configuration.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

// Configure

func (r *QConnectAIPromptVersionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *aws.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.config = cfg
}

// Create

func (r *QConnectAIPromptVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data QConnectAIPromptVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := qconnect.NewFromConfig(r.config)

	in := &qconnect.CreateAIPromptVersionInput{
		AssistantId: aws.String(data.AssistantId.ValueString()),
		AiPromptId:  aws.String(data.AiPromptId.ValueString()),
	}
	if !data.ModifiedTimeSeconds.IsNull() && !data.ModifiedTimeSeconds.IsUnknown() {
		t := time.Unix(data.ModifiedTimeSeconds.ValueInt64(), 0)
		in.ModifiedTime = &t
	}

	out, err := conn.CreateAIPromptVersion(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Q in Connect AI Prompt version", err.Error())
		return
	}

	tflog.Trace(ctx, "Created qconnect AI prompt version")

	versionNumber := aws.ToInt64(out.VersionNumber)
	data.VersionNumber = types.Int64Value(versionNumber)

	promptArn := ""
	if out.AiPrompt != nil {
		promptArn = aws.ToString(out.AiPrompt.AiPromptArn)
	}
	data.PromptArnWithVersion = types.StringValue(fmt.Sprintf("%s:%d", promptArn, versionNumber))
	data.AiPromptIdWithVersion = types.StringValue(fmt.Sprintf("%s:%d", data.AiPromptId.ValueString(), versionNumber))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read

func (r *QConnectAIPromptVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data QConnectAIPromptVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := qconnect.NewFromConfig(r.config)

	versionedId := fmt.Sprintf("%s:%d", data.AiPromptId.ValueString(), data.VersionNumber.ValueInt64())

	out, err := conn.GetAIPrompt(ctx, &qconnect.GetAIPromptInput{
		AssistantId: aws.String(data.AssistantId.ValueString()),
		AiPromptId:  aws.String(versionedId),
	})
	if err != nil {
		var nf *qconnecttypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Q in Connect AI Prompt version", err.Error())
		return
	}
	if out == nil || out.AiPrompt == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	versionNumber := data.VersionNumber.ValueInt64()
	if out.VersionNumber != nil {
		versionNumber = aws.ToInt64(out.VersionNumber)
	}
	data.VersionNumber = types.Int64Value(versionNumber)
	data.PromptArnWithVersion = types.StringValue(fmt.Sprintf("%s:%d", aws.ToString(out.AiPrompt.AiPromptArn), versionNumber))
	data.AiPromptIdWithVersion = types.StringValue(fmt.Sprintf("%s:%d", data.AiPromptId.ValueString(), versionNumber))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update - All user-facing attributes are ForceNew or Computed

func (r *QConnectAIPromptVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"awsext_qconnect_ai_prompt_version does not support in-place updates. All changes force resource replacement.",
	)
}

// Delete

func (r *QConnectAIPromptVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data QConnectAIPromptVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := qconnect.NewFromConfig(r.config)

	_, err := conn.DeleteAIPromptVersion(ctx, &qconnect.DeleteAIPromptVersionInput{
		AssistantId:   aws.String(data.AssistantId.ValueString()),
		AiPromptId:    aws.String(data.AiPromptId.ValueString()),
		VersionNumber: aws.Int64(data.VersionNumber.ValueInt64()),
	})
	if err != nil {
		var nf *qconnecttypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting Q in Connect AI Prompt version", err.Error())
	}
}

// ImportState

func (r *QConnectAIPromptVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected <assistant_id>/<ai_prompt_id>/<version_number>",
		)
		return
	}

	versionNumber, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid version_number in import ID",
			fmt.Sprintf("Could not parse %q as an integer: %s", parts[2], err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("assistant_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ai_prompt_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version_number"), versionNumber)...)
}
