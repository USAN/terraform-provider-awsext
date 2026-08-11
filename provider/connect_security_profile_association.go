package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	conntypes "github.com/aws/aws-sdk-go-v2/service/connect/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ConnectSecurityProfileAssociationResource{}
var _ resource.ResourceWithImportState = &ConnectSecurityProfileAssociationResource{}

func NewConnectSecurityProfileAssociationResource() resource.Resource {
	return &ConnectSecurityProfileAssociationResource{}
}

type ConnectSecurityProfileAssociationResource struct {
	config aws.Config
}

type ConnectSecurityProfileAssociationResourceModel struct {
	InstanceID         types.String `tfsdk:"instance_id"`
	SecurityProfileID  types.String `tfsdk:"security_profile_id"`
	EntityType         types.String `tfsdk:"entity_type"`
	EntityArn          types.String `tfsdk:"entity_arn"`
	ID                 types.String `tfsdk:"id"`
}

// securityProfileAssociationScopes returns the three ARN scopes Amazon Connect
// tracks independently for an entity's security-profile association: the base
// ARN, its ":$LATEST" alias, and its ":$SAVED" alias. Associating one does not
// associate the others (see AssociateSecurityProfiles / ListEntitySecurityProfiles
// docs), so every Create/Read/Delete must fan out across all three.
func securityProfileAssociationScopes(entityArn string) []string {
	return []string{
		entityArn,
		entityArn + ":$LATEST",
		entityArn + ":$SAVED",
	}
}

// connectSecurityProfileAssociationID builds the synthetic resource ID used for
// import: "<instance_id>/<security_profile_id>/<entity_arn>".
func connectSecurityProfileAssociationID(instanceID, securityProfileID, entityArn string) string {
	return instanceID + "/" + securityProfileID + "/" + entityArn
}

func isResourceConflict(err error) bool {
	var confErr *conntypes.ResourceConflictException
	return errors.As(err, &confErr)
}

func isResourceNotFound(err error) bool {
	var nfErr *conntypes.ResourceNotFoundException
	return errors.As(err, &nfErr)
}

// -------------------------------------------------------------------
// Metadata / Schema
// -------------------------------------------------------------------

func (r *ConnectSecurityProfileAssociationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connect_security_profile_association"
}

func (r *ConnectSecurityProfileAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Associates an Amazon Connect Security Profile with an entity (currently only " +
			"`AI_AGENT` is supported by the underlying API) across all three ARN scopes Amazon Connect tracks " +
			"independently: the base ARN, `:$LATEST`, and `:$SAVED`. Fills the gap left by the official " +
			"`aws_connect_security_profile` resource and by `awsext_connect_security_profile`, neither of which " +
			"attaches the profile to anything — an AI Agent otherwise fails every MCP tool call with " +
			"\"Insufficient\" permission.",

		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "The identifier of the Amazon Connect instance. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_profile_id": schema.StringAttribute{
				Required:    true,
				Description: "The identifier of the Connect Security Profile to associate. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_type": schema.StringAttribute{
				Required:    true,
				Description: "The entity type to associate the security profile with. Only \"AI_AGENT\" is supported by the underlying API today. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_arn": schema.StringAttribute{
				Required:    true,
				Description: "The base ARN of the entity (no version/alias suffix) — the version and $LATEST/$SAVED scopes are associated automatically alongside it. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic identifier: \"<instance_id>/<security_profile_id>/<entity_arn>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// -------------------------------------------------------------------
// Configure
// -------------------------------------------------------------------

func (r *ConnectSecurityProfileAssociationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// -------------------------------------------------------------------
// CRUD Stubs (Task 2)
// -------------------------------------------------------------------

func (r *ConnectSecurityProfileAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	panic("not implemented")
}

func (r *ConnectSecurityProfileAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	panic("not implemented")
}

func (r *ConnectSecurityProfileAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("not implemented")
}

func (r *ConnectSecurityProfileAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	panic("not implemented")
}

func (r *ConnectSecurityProfileAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	panic("not implemented")
}
