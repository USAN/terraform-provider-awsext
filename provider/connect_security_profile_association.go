package provider

import (
	"context"
	"errors"
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
	InstanceID        types.String `tfsdk:"instance_id"`
	SecurityProfileID types.String `tfsdk:"security_profile_id"`
	EntityType        types.String `tfsdk:"entity_type"`
	EntityArn         types.String `tfsdk:"entity_arn"`
	ID                types.String `tfsdk:"id"`
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

// disassociateSecurityProfilesFromEntity lists every security profile
// currently associated with entityArn (via the documented, structured
// ListEntitySecurityProfiles lookup — entity-to-profiles is the only
// direction AWS exposes; there is no reverse profile-to-entities API) and
// disassociates each one. Used to clear a specific ARN scope (e.g. a
// numbered AI Agent version) that awsext_connect_security_profile_association
// never manages, since it only fans out to the base/$LATEST/$SAVED scopes.
func disassociateSecurityProfilesFromEntity(ctx context.Context, conn *connect.Client, instanceID, entityArn string) error {
	paginator := connect.NewListEntitySecurityProfilesPaginator(conn, &connect.ListEntitySecurityProfilesInput{
		InstanceId: aws.String(instanceID),
		EntityType: conntypes.EntityTypeAiAgent,
		EntityArn:  aws.String(entityArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if isResourceNotFound(err) {
				return nil
			}
			return fmt.Errorf("listing security profiles for %s: %w", entityArn, err)
		}
		for _, sp := range page.SecurityProfiles {
			_, err := conn.DisassociateSecurityProfiles(ctx, &connect.DisassociateSecurityProfilesInput{
				InstanceId: aws.String(instanceID),
				EntityArn:  aws.String(entityArn),
				EntityType: conntypes.EntityTypeAiAgent,
				SecurityProfiles: []conntypes.SecurityProfileItem{
					{Id: sp.Id},
				},
			})
			if err != nil && !isResourceNotFound(err) {
				return fmt.Errorf("disassociating security profile %s from %s: %w", aws.ToString(sp.Id), entityArn, err)
			}
		}
	}
	return nil
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
// Create
// -------------------------------------------------------------------

func (r *ConnectSecurityProfileAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConnectSecurityProfileAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(r.config)

	for _, scopeArn := range securityProfileAssociationScopes(data.EntityArn.ValueString()) {
		_, err := conn.AssociateSecurityProfiles(ctx, &connect.AssociateSecurityProfilesInput{
			InstanceId: aws.String(data.InstanceID.ValueString()),
			EntityArn:  aws.String(scopeArn),
			EntityType: conntypes.EntityType(data.EntityType.ValueString()),
			SecurityProfiles: []conntypes.SecurityProfileItem{
				{Id: aws.String(data.SecurityProfileID.ValueString())},
			},
		})
		if err != nil && !isResourceConflict(err) {
			resp.Diagnostics.AddError(
				"Error associating Connect Security Profile",
				fmt.Sprintf("Could not associate security profile %s with entity %s: %s", data.SecurityProfileID.ValueString(), scopeArn, err),
			)
			return
		}
	}

	data.ID = types.StringValue(connectSecurityProfileAssociationID(
		data.InstanceID.ValueString(), data.SecurityProfileID.ValueString(), data.EntityArn.ValueString(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Read
// -------------------------------------------------------------------

func (r *ConnectSecurityProfileAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConnectSecurityProfileAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(r.config)

	for _, scopeArn := range securityProfileAssociationScopes(data.EntityArn.ValueString()) {
		associated, err := r.scopeHasSecurityProfile(ctx, conn, data.InstanceID.ValueString(), data.EntityType.ValueString(), scopeArn, data.SecurityProfileID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading Connect Security Profile Association",
				fmt.Sprintf("Could not list entity security profiles for %s: %s", scopeArn, err),
			)
			return
		}
		if !associated {
			// All-or-nothing: if any of the 3 scopes has drifted, drop the
			// whole resource from state so the next apply recreates all 3
			// scopes cleanly rather than tracking partial association state.
			resp.State.RemoveResource(ctx)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// scopeHasSecurityProfile pages through ListEntitySecurityProfiles for one ARN
// scope and reports whether securityProfileID is among the associated profiles.
func (r *ConnectSecurityProfileAssociationResource) scopeHasSecurityProfile(ctx context.Context, conn *connect.Client, instanceID, entityType, entityArn, securityProfileID string) (bool, error) {
	paginator := connect.NewListEntitySecurityProfilesPaginator(conn, &connect.ListEntitySecurityProfilesInput{
		InstanceId: aws.String(instanceID),
		EntityType: conntypes.EntityType(entityType),
		EntityArn:  aws.String(entityArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return false, err
		}
		for _, sp := range page.SecurityProfiles {
			if aws.ToString(sp.Id) == securityProfileID {
				return true, nil
			}
		}
	}
	return false, nil
}

// -------------------------------------------------------------------
// Update
// -------------------------------------------------------------------

// Update is never expected to perform any in-place mutation: every attribute
// in the schema either forces replacement (instance_id, security_profile_id,
// entity_type, entity_arn) or is Computed with UseStateForUnknown (id). The
// terraform-plugin-framework resource.Resource interface still requires an
// Update method, and the framework calls it whenever a plan has no
// RequiresReplace diffs but is not entirely a no-op (e.g. after a
// provider-driven state refresh). The minimal correct implementation is to
// carry the plan straight into state.
func (r *ConnectSecurityProfileAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConnectSecurityProfileAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Delete
// -------------------------------------------------------------------

func (r *ConnectSecurityProfileAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConnectSecurityProfileAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(r.config)

	for _, scopeArn := range securityProfileAssociationScopes(data.EntityArn.ValueString()) {
		_, err := conn.DisassociateSecurityProfiles(ctx, &connect.DisassociateSecurityProfilesInput{
			InstanceId: aws.String(data.InstanceID.ValueString()),
			EntityArn:  aws.String(scopeArn),
			EntityType: conntypes.EntityType(data.EntityType.ValueString()),
			SecurityProfiles: []conntypes.SecurityProfileItem{
				{Id: aws.String(data.SecurityProfileID.ValueString())},
			},
		})
		if err != nil && !isResourceNotFound(err) {
			resp.Diagnostics.AddError(
				"Error disassociating Connect Security Profile",
				fmt.Sprintf("Could not disassociate security profile %s from entity %s: %s", data.SecurityProfileID.ValueString(), scopeArn, err),
			)
			return
		}
	}
}

// -------------------------------------------------------------------
// ImportState
// -------------------------------------------------------------------

// ImportState accepts the import ID in the format
// <instance_id>/<security_profile_id>/<entity_arn>.
//
// Import command:
//
//	terraform import awsext_connect_security_profile_association.example <instance_id>/<security_profile_id>/<entity_arn>
func (r *ConnectSecurityProfileAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in the format <instance_id>/<security_profile_id>/<entity_arn>, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("security_profile_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entity_arn"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entity_type"), "AI_AGENT")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
