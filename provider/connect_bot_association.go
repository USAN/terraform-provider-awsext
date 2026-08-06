package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	conntypes "github.com/aws/aws-sdk-go-v2/service/connect/types"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ConnectBotAssociationResource{}
var _ resource.ResourceWithImportState = &ConnectBotAssociationResource{}

func NewConnectBotAssociationResource() resource.Resource {
	return &ConnectBotAssociationResource{}
}

type ConnectBotAssociationResource struct {
	config aws.Config
}

// -------------------------------------------------------------------
// Model
// -------------------------------------------------------------------

type ConnectBotAssociationResourceModel struct {
	InstanceID types.String `tfsdk:"instance_id"`
	AliasArn   types.String `tfsdk:"alias_arn"`
}

// -------------------------------------------------------------------
// Metadata / Schema
// -------------------------------------------------------------------

func (r *ConnectBotAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connect_bot_association"
}

func (r *ConnectBotAssociationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Associates an Amazon Lex V2 bot alias with an Amazon Connect instance " +
			"(connect:AssociateBot / DisassociateBot). Fills the gap left by aws_connect_bot_association, " +
			"which only supports Lex V1.",

		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "The identifier of the Amazon Connect instance. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alias_arn": schema.StringAttribute{
				Required:    true,
				Description: "The ARN of the Lex V2 bot alias to associate. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// -------------------------------------------------------------------
// Configure
// -------------------------------------------------------------------

func (r *ConnectBotAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// -------------------------------------------------------------------
// Create
// -------------------------------------------------------------------

func (r *ConnectBotAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConnectBotAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(r.config)

	_, err := conn.AssociateBot(ctx, &connect.AssociateBotInput{
		InstanceId: aws.String(data.InstanceID.ValueString()),
		LexV2Bot: &conntypes.LexV2Bot{
			AliasArn: aws.String(data.AliasArn.ValueString()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error associating Lex V2 bot with Connect instance", err.Error())
		return
	}

	tflog.Trace(ctx, "associated Lex V2 bot with Connect instance",
		map[string]any{"instance_id": data.InstanceID.ValueString(), "alias_arn": data.AliasArn.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Read
// -------------------------------------------------------------------

func (r *ConnectBotAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConnectBotAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(r.config)

	found, err := r.findAssociation(ctx, conn, data.InstanceID.ValueString(), data.AliasArn.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Connect bot association", err.Error())
		return
	}

	if !found {
		tflog.Warn(ctx, "Connect bot association not found, removing from state",
			map[string]any{"alias_arn": data.AliasArn.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Update — no in-place updates; all attributes force replacement
// -------------------------------------------------------------------

func (r *ConnectBotAssociationResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// -------------------------------------------------------------------
// Delete
// -------------------------------------------------------------------

func (r *ConnectBotAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConnectBotAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(r.config)

	_, err := conn.DisassociateBot(ctx, &connect.DisassociateBotInput{
		InstanceId: aws.String(data.InstanceID.ValueString()),
		LexV2Bot: &conntypes.LexV2Bot{
			AliasArn: aws.String(data.AliasArn.ValueString()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error disassociating Lex V2 bot from Connect instance", err.Error())
	}
}

// -------------------------------------------------------------------
// ImportState
// -------------------------------------------------------------------

// ImportState accepts the import ID in the format <instance_id>/<alias_arn>.
// Because alias ARNs contain slashes, we split only on the first '/'.
func (r *ConnectBotAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be <instance_id>/<alias_arn>, got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("alias_arn"), parts[1])...)
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func (r *ConnectBotAssociationResource) findAssociation(ctx context.Context, conn *connect.Client, instanceID, aliasArn string) (bool, error) {
	var nextToken *string
	for {
		out, err := conn.ListBots(ctx, &connect.ListBotsInput{
			InstanceId: aws.String(instanceID),
			LexVersion: conntypes.LexVersionV2,
			NextToken:  nextToken,
		})
		if err != nil {
			return false, fmt.Errorf("ListBots: %w", err)
		}
		for _, entry := range out.LexBots {
			if entry.LexV2Bot != nil && aws.ToString(entry.LexV2Bot.AliasArn) == aliasArn {
				return true, nil
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return false, nil
}
