package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
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

var _ resource.Resource = &QConnectAIAgentVersionResource{}
var _ resource.ResourceWithImportState = &QConnectAIAgentVersionResource{}

// versionedAIAgentArn qualifies agentArn with ":<versionNumber>", stripping
// any existing numeric suffix first. CreateAIAgentVersion returns the bare
// entity ARN (needs the suffix appended), but GetAIAgent called with an
// already-versioned ai_agent_id returns the ARN already qualified — treating
// both the same way (as this code used to) doubles the suffix into
// ".../<id>:1:1", corrupting state on every refresh and breaking any lookup
// against the real ARN (confirmed against prod-mock 2026-08-27:
// ListEntitySecurityProfiles rejected the doubled ARN during Delete's
// disassociate step).
func versionedAIAgentArn(agentArn string, versionNumber int64) string {
	base := strings.TrimSuffix(agentArn, fmt.Sprintf(":%d", versionNumber))
	return fmt.Sprintf("%s:%d", base, versionNumber)
}

func NewQConnectAIAgentVersionResource() resource.Resource {
	return &QConnectAIAgentVersionResource{}
}

type QConnectAIAgentVersionResource struct{ config aws.Config }

type QConnectAIAgentVersionResourceModel struct {
	AssistantId           types.String `tfsdk:"assistant_id"`
	AiAgentId             types.String `tfsdk:"ai_agent_id"`
	ModifiedTimeSeconds   types.Int64  `tfsdk:"modified_time_seconds"`
	ConnectInstanceID     types.String `tfsdk:"connect_instance_id"`
	VersionNumber         types.Int64  `tfsdk:"version_number"`
	AIAgentARNWithVersion types.String `tfsdk:"ai_agent_arn_with_version"`
	AIAgentIdWithVersion  types.String `tfsdk:"ai_agent_id_with_version"`
}

// Metadata

func (r *QConnectAIAgentVersionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qconnect_ai_agent_version"
}

// Schema

func (r *QConnectAIAgentVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a version snapshot of an Amazon Q in Connect AI Agent (qconnect:CreateAIAgentVersion). Versions are immutable; all inputs are ForceNew. No update API exists.",
		Attributes: map[string]schema.Attribute{
			"assistant_id": schema.StringAttribute{
				Required:      true,
				Description:   "Identifier of the parent Amazon Q in Connect Assistant.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ai_agent_id": schema.StringAttribute{
				Required:      true,
				Description:   "Identifier of the AI Agent to version.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"modified_time_seconds": schema.Int64Attribute{
				Optional:      true,
				Description:   "Unix epoch seconds of the last-known modification time of the AI agent. When set, the API rejects the version create if the AI agent has been modified more recently, preventing accidental version creation on stale configs.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"connect_instance_id": schema.StringAttribute{
				Optional:      true,
				Description:   "Amazon Connect instance ID. When set, Delete disassociates this version's numbered ARN from any Connect Security Profile still holding it (AssociateSecurityProfiles fans a base entity_arn out to 3 scopes, none of which is a numbered version — AWS never disassociates the numbered scope itself when the version is deleted, which otherwise permanently blocks DeleteSecurityProfile). Omit if this AI Agent's security profile associations aren't managed via awsext_connect_security_profile_association.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"version_number": schema.Int64Attribute{
				Computed:      true,
				Description:   "Service-assigned version number of the AI Agent version.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"ai_agent_arn_with_version": schema.StringAttribute{
				Computed:      true,
				Description:   "Full ARN of the AI Agent qualified with the version number",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ai_agent_id_with_version": schema.StringAttribute{
				Computed:      true,
				Description:   "AI Agent ID qualified with the version number (<ai_agent_id>:<version_number>). Use this as the versioned agent reference in Connect session configurations.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

// Configure

func (r *QConnectAIAgentVersionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *QConnectAIAgentVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data QConnectAIAgentVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := qconnect.NewFromConfig(r.config)

	in := &qconnect.CreateAIAgentVersionInput{
		AssistantId: aws.String(data.AssistantId.ValueString()),
		AiAgentId:   aws.String(data.AiAgentId.ValueString()),
	}
	if !data.ModifiedTimeSeconds.IsNull() && !data.ModifiedTimeSeconds.IsUnknown() {
		t := time.Unix(data.ModifiedTimeSeconds.ValueInt64(), 0)
		in.ModifiedTime = &t
	}

	out, err := conn.CreateAIAgentVersion(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Q in Connect AI Agent version", err.Error())
		return
	}

	tflog.Trace(ctx, "created qconnect AI Agent version")

	versionNumber := aws.ToInt64(out.VersionNumber)
	data.VersionNumber = types.Int64Value(versionNumber)

	agentArn := ""
	if out.AiAgent != nil {
		agentArn = aws.ToString(out.AiAgent.AiAgentArn)
	}
	data.AIAgentARNWithVersion = types.StringValue(versionedAIAgentArn(agentArn, versionNumber))
	data.AIAgentIdWithVersion = types.StringValue(fmt.Sprintf("%s:%d", data.AiAgentId.ValueString(), versionNumber))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read

func (r *QConnectAIAgentVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data QConnectAIAgentVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := qconnect.NewFromConfig(r.config)

	versionedId := fmt.Sprintf("%s:%d", data.AiAgentId.ValueString(), data.VersionNumber.ValueInt64())

	out, err := conn.GetAIAgent(ctx, &qconnect.GetAIAgentInput{
		AssistantId: aws.String(data.AssistantId.ValueString()),
		AiAgentId:   aws.String(versionedId),
	})
	if err != nil {
		var nf *qconnecttypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Q in Connect AI Agent version", err.Error())
		return
	}
	if out == nil || out.AiAgent == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	versionNumber := data.VersionNumber.ValueInt64()
	if out.VersionNumber != nil {
		versionNumber = aws.ToInt64(out.VersionNumber)
	}
	data.VersionNumber = types.Int64Value(versionNumber)
	data.AIAgentARNWithVersion = types.StringValue(
		versionedAIAgentArn(aws.ToString(out.AiAgent.AiAgentArn), versionNumber),
	)
	data.AIAgentIdWithVersion = types.StringValue(
		fmt.Sprintf("%s:%d", data.AiAgentId.ValueString(), versionNumber),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update

func (r *QConnectAIAgentVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"awsext_qconnect_ai_agent_version does not support in-place updates. All changes force resource replacement.",
	)
}

// Delete

func (r *QConnectAIAgentVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data QConnectAIAgentVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ConnectInstanceID.IsNull() && !data.ConnectInstanceID.IsUnknown() {
		if err := disassociateSecurityProfilesFromEntity(
			ctx, connect.NewFromConfig(r.config),
			data.ConnectInstanceID.ValueString(), data.AIAgentARNWithVersion.ValueString(),
		); err != nil {
			resp.Diagnostics.AddError(
				"Error disassociating Connect Security Profiles from AI Agent version",
				fmt.Sprintf("Could not clear security profile associations for %s before delete: %s",
					data.AIAgentARNWithVersion.ValueString(), err),
			)
			return
		}
	}

	conn := qconnect.NewFromConfig(r.config)

	_, err := conn.DeleteAIAgentVersion(ctx, &qconnect.DeleteAIAgentVersionInput{
		AssistantId:   aws.String(data.AssistantId.ValueString()),
		AiAgentId:     aws.String(data.AiAgentId.ValueString()),
		VersionNumber: aws.Int64(data.VersionNumber.ValueInt64()),
	})
	if err != nil {
		var nf *qconnecttypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting Q in Connect AI Agent version", err.Error())
	}
}

// ImportState

func (r *QConnectAIAgentVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected <assistant_id>/<ai_agent_id>/<version_number>",
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ai_agent_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version_number"), versionNumber)...)
}
