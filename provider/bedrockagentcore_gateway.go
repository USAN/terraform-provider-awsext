// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	bactypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &BedrockAgentCoreGatewayResource{}
var _ resource.ResourceWithImportState = &BedrockAgentCoreGatewayResource{}

func NewBedrockAgentCoreGatewayResource() resource.Resource {
	return &BedrockAgentCoreGatewayResource{}
}

type BedrockAgentCoreGatewayResource struct{ config aws.Config }

// -------------------------------------------------------------------
// Model
// -------------------------------------------------------------------

type BedrockAgentCoreGatewayResourceModel struct {
	GatewayId               types.String                            `tfsdk:"gateway_id"`
	GatewayArn              types.String                            `tfsdk:"gateway_arn"`
	GatewayUrl              types.String                            `tfsdk:"gateway_url"`
	Status                  types.String                            `tfsdk:"status"`
	Name                    types.String                            `tfsdk:"name"`
	Description             types.String                            `tfsdk:"description"`
	RoleArn                 types.String                            `tfsdk:"role_arn"`
	ProtocolType            types.String                            `tfsdk:"protocol_type"`
	AuthorizerType          types.String                            `tfsdk:"authorizer_type"`
	ExceptionLevel          types.String                            `tfsdk:"exception_level"`
	KmsKeyArn               types.String                            `tfsdk:"kms_key_arn"`
	CustomJwtAuthorizer     *BedrockAgentCoreCustomJWTAuthorizerModel `tfsdk:"custom_jwt_authorizer"`
	AllowedAudience         types.List                              `tfsdk:"allowed_audience"`
}

type BedrockAgentCoreCustomJWTAuthorizerModel struct {
	DiscoveryUrl types.String `tfsdk:"discovery_url"`
}

// -------------------------------------------------------------------
// Metadata / Schema
// -------------------------------------------------------------------

func (r *BedrockAgentCoreGatewayResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bedrockagentcore_gateway"
}

func (r *BedrockAgentCoreGatewayResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages an Amazon Bedrock AgentCore Gateway " +
			"(bedrock-agentcore-control:CreateGateway / GetGateway / UpdateGateway / DeleteGateway). " +
			"The gateway is created with a bootstrap audience, then the resource immediately calls " +
			"UpdateGateway to set allowed_audience to the gateway's own gateway_id. This eliminates " +
			"the chicken-and-egg problem where the AppIntegration namespace and JWT audience must " +
			"both equal a value the gateway itself produces.",

		Attributes: map[string]schema.Attribute{
			"gateway_id": schema.StringAttribute{
				Computed:    true,
				Description: "Service-assigned unique identifier of the gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway_arn": schema.StringAttribute{
				Computed:    true,
				Description: "Amazon Resource Name (ARN) of the gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway_url": schema.StringAttribute{
				Computed:    true,
				Description: "Endpoint URL for invoking the gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the gateway (e.g. READY, CREATING, UPDATING).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the gateway. Must be unique within the account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the gateway. Updatable in place.",
			},
			"role_arn": schema.StringAttribute{
				Required:    true,
				Description: "IAM role ARN granting the gateway permission to invoke target Lambdas, read S3, etc. Updatable in place.",
			},
			"protocol_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Protocol type for the gateway. Currently only `MCP` is supported. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("MCP"),
				},
			},
			"authorizer_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Authorizer type. Only `CUSTOM_JWT` is supported by this resource; choose another type by using aws_bedrockagentcore_gateway. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("CUSTOM_JWT"),
				},
			},
			"exception_level": schema.StringAttribute{
				Optional:    true,
				Description: "Verbosity of error messages returned to callers. Set to `DEBUG` for granular messages. Updatable in place.",
				Validators: []validator.String{
					stringvalidator.OneOf("DEBUG"),
				},
			},
			"kms_key_arn": schema.StringAttribute{
				Optional:    true,
				Description: "Customer-managed KMS key ARN used to encrypt gateway data. Updatable in place.",
			},
			"custom_jwt_authorizer": schema.SingleNestedAttribute{
				Required:    true,
				Description: "JWT authorizer configuration. The `allowed_audience` is managed automatically by this resource and set to the gateway's own gateway_id after creation; do not configure it here.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"discovery_url": schema.StringAttribute{
						Required:    true,
						Description: "OpenID Connect discovery URL of the identity provider that issues JWTs the gateway will accept.",
					},
				},
			},
			"allowed_audience": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Audience values accepted on incoming JWTs. Set by this resource to [gateway_id] after create/update.",
				PlanModifiers: []planmodifier.List{},
			},
		},
	}
}

// -------------------------------------------------------------------
// Configure
// -------------------------------------------------------------------

func (r *BedrockAgentCoreGatewayResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected aws.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.config = cfg
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

// bootstrapAudience is the placeholder audience used on the initial CreateGateway
// call. It is immediately replaced by [gateway_id] via UpdateGateway. The value
// is never visible to callers in state because the resource always reads back
// the post-update audience.
const bootstrapAudience = "__awsext_bootstrap__"

func protocolTypeOrDefault(s types.String) bactypes.GatewayProtocolType {
	if s.IsNull() || s.IsUnknown() || s.ValueString() == "" {
		return bactypes.GatewayProtocolTypeMcp
	}
	return bactypes.GatewayProtocolType(s.ValueString())
}

func authorizerTypeOrDefault(s types.String) bactypes.AuthorizerType {
	if s.IsNull() || s.IsUnknown() || s.ValueString() == "" {
		return bactypes.AuthorizerTypeCustomJwt
	}
	return bactypes.AuthorizerType(s.ValueString())
}

// -------------------------------------------------------------------
// Create
// -------------------------------------------------------------------

func (r *BedrockAgentCoreGatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BedrockAgentCoreGatewayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.CustomJwtAuthorizer == nil {
		resp.Diagnostics.AddError(
			"Missing custom_jwt_authorizer",
			"custom_jwt_authorizer is required but was not provided.",
		)
		return
	}

	conn := bedrockagentcorecontrol.NewFromConfig(r.config)

	discoveryUrl := data.CustomJwtAuthorizer.DiscoveryUrl.ValueString()
	authCfg := &bactypes.AuthorizerConfigurationMemberCustomJWTAuthorizer{
		Value: bactypes.CustomJWTAuthorizerConfiguration{
			DiscoveryUrl:    aws.String(discoveryUrl),
			AllowedAudience: []string{bootstrapAudience},
		},
	}

	in := &bedrockagentcorecontrol.CreateGatewayInput{
		Name:                    aws.String(data.Name.ValueString()),
		RoleArn:                 aws.String(data.RoleArn.ValueString()),
		ProtocolType:            protocolTypeOrDefault(data.ProtocolType),
		AuthorizerType:          authorizerTypeOrDefault(data.AuthorizerType),
		AuthorizerConfiguration: authCfg,
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() && data.Description.ValueString() != "" {
		in.Description = aws.String(data.Description.ValueString())
	}
	if !data.ExceptionLevel.IsNull() && !data.ExceptionLevel.IsUnknown() && data.ExceptionLevel.ValueString() != "" {
		in.ExceptionLevel = bactypes.ExceptionLevel(data.ExceptionLevel.ValueString())
	}
	if !data.KmsKeyArn.IsNull() && !data.KmsKeyArn.IsUnknown() && data.KmsKeyArn.ValueString() != "" {
		in.KmsKeyArn = aws.String(data.KmsKeyArn.ValueString())
	}

	out, err := conn.CreateGateway(ctx, in)

	var gatewayId, gatewayArn, gatewayUrl string
	var gatewayStatus string

	if err != nil {
		var ce *bactypes.ConflictException
		if !errors.As(err, &ce) {
			resp.Diagnostics.AddError("Error creating AgentCore Gateway", err.Error())
			return
		}
		// Gateway already exists — look it up by name and adopt it.
		existingId, lookupErr := findGatewayIdByName(ctx, conn, data.Name.ValueString())
		if lookupErr != nil {
			resp.Diagnostics.AddError(
				"Error adopting existing AgentCore Gateway",
				fmt.Sprintf("A gateway named %q already exists but could not be located: %s", data.Name.ValueString(), lookupErr),
			)
			return
		}
		getOut, getErr := conn.GetGateway(ctx, &bedrockagentcorecontrol.GetGatewayInput{
			GatewayIdentifier: aws.String(existingId),
		})
		if getErr != nil {
			resp.Diagnostics.AddError(
				"Error reading existing AgentCore Gateway",
				fmt.Sprintf("Located gateway %s by name but could not read it: %s", existingId, getErr),
			)
			return
		}
		gatewayId = aws.ToString(getOut.GatewayId)
		gatewayArn = aws.ToString(getOut.GatewayArn)
		gatewayUrl = aws.ToString(getOut.GatewayUrl)
		gatewayStatus = string(getOut.Status)
	} else {
		gatewayId = aws.ToString(out.GatewayId)
		gatewayArn = aws.ToString(out.GatewayArn)
		gatewayUrl = aws.ToString(out.GatewayUrl)
		gatewayStatus = string(out.Status)
	}

	// Save partial state immediately. The gateway exists in AWS now; if anything
	// below fails, terraform must still track it so the next apply can either
	// finish the audience patch via Update or destroy the gateway via Delete.
	// Without this, a failed UpdateGateway leaves an orphan that blocks the
	// next CreateGateway with a name ConflictException.
	r.flushOutputToState(ctx, &data, gatewayId, gatewayArn, gatewayUrl, gatewayStatus, discoveryUrl, []string{bootstrapAudience})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Immediately PATCH allowed_audience to [gateway_id]. This is the whole reason
	// this resource exists: AWS Connect's AppIntegration namespace must match the
	// gateway id, and the JWT audience must match the namespace, so the gateway
	// has to accept its own id as audience — a value not known until after Create.
	upIn := &bedrockagentcorecontrol.UpdateGatewayInput{
		GatewayIdentifier: aws.String(gatewayId),
		Name:              in.Name,
		RoleArn:           in.RoleArn,
		ProtocolType:      in.ProtocolType,
		AuthorizerType:    in.AuthorizerType,
		AuthorizerConfiguration: &bactypes.AuthorizerConfigurationMemberCustomJWTAuthorizer{
			Value: bactypes.CustomJWTAuthorizerConfiguration{
				DiscoveryUrl:    aws.String(discoveryUrl),
				AllowedAudience: []string{gatewayId},
			},
		},
		Description:    in.Description,
		ExceptionLevel: in.ExceptionLevel,
		KmsKeyArn:      in.KmsKeyArn,
	}

	// CreateGateway returns while the gateway is still in CREATING state. The
	// follow-up UpdateGateway must wait until it reaches READY, otherwise AWS
	// returns ValidationException: "UpdateGateway operation can't be performed
	// on gateway when it is in Creating state". Retry on that specific error.
	upOut, err := retryUpdateGatewayUntilReady(ctx, conn, upIn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting self-audience on AgentCore Gateway",
			fmt.Sprintf("Gateway %s was created and saved to state, but the follow-up UpdateGateway call to set allowed_audience=[%s] failed: %s. The gateway exists in AWS with allowed_audience=[%s] and will not accept JWTs until corrected. Re-run `terraform apply` to retry the audience patch via the Update path.", gatewayId, gatewayId, err, bootstrapAudience),
		)
		return
	}

	r.flushOutputToState(ctx, &data, gatewayId, gatewayArn, aws.ToString(upOut.GatewayUrl), string(upOut.Status), discoveryUrl, []string{gatewayId})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Read
// -------------------------------------------------------------------

func (r *BedrockAgentCoreGatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BedrockAgentCoreGatewayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := bedrockagentcorecontrol.NewFromConfig(r.config)

	out, err := conn.GetGateway(ctx, &bedrockagentcorecontrol.GetGatewayInput{
		GatewayIdentifier: aws.String(data.GatewayId.ValueString()),
	})
	if err != nil {
		var nf *bactypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading AgentCore Gateway",
			fmt.Sprintf("Could not read gateway %s: %s", data.GatewayId.ValueString(), err))
		return
	}
	if out == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	discoveryUrl := ""
	var audience []string
	if jwt, ok := out.AuthorizerConfiguration.(*bactypes.AuthorizerConfigurationMemberCustomJWTAuthorizer); ok && jwt != nil {
		discoveryUrl = aws.ToString(jwt.Value.DiscoveryUrl)
		audience = jwt.Value.AllowedAudience
	}

	r.flushOutputToState(ctx, &data,
		aws.ToString(out.GatewayId),
		aws.ToString(out.GatewayArn),
		aws.ToString(out.GatewayUrl),
		string(out.Status),
		discoveryUrl,
		audience,
	)
	data.Name = types.StringValue(aws.ToString(out.Name))
	data.RoleArn = types.StringValue(aws.ToString(out.RoleArn))
	data.ProtocolType = types.StringValue(string(out.ProtocolType))
	data.AuthorizerType = types.StringValue(string(out.AuthorizerType))

	if out.Description != nil {
		data.Description = types.StringValue(aws.ToString(out.Description))
	} else {
		data.Description = types.StringNull()
	}
	if out.ExceptionLevel != "" {
		data.ExceptionLevel = types.StringValue(string(out.ExceptionLevel))
	} else {
		data.ExceptionLevel = types.StringNull()
	}
	if out.KmsKeyArn != nil {
		data.KmsKeyArn = types.StringValue(aws.ToString(out.KmsKeyArn))
	} else {
		data.KmsKeyArn = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Update
// -------------------------------------------------------------------

func (r *BedrockAgentCoreGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BedrockAgentCoreGatewayResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.CustomJwtAuthorizer == nil {
		resp.Diagnostics.AddError(
			"Missing custom_jwt_authorizer",
			"custom_jwt_authorizer is required but was not provided.",
		)
		return
	}

	conn := bedrockagentcorecontrol.NewFromConfig(r.config)

	gatewayId := state.GatewayId.ValueString()
	discoveryUrl := plan.CustomJwtAuthorizer.DiscoveryUrl.ValueString()

	in := &bedrockagentcorecontrol.UpdateGatewayInput{
		GatewayIdentifier: aws.String(gatewayId),
		Name:              aws.String(plan.Name.ValueString()),
		RoleArn:           aws.String(plan.RoleArn.ValueString()),
		ProtocolType:      protocolTypeOrDefault(plan.ProtocolType),
		AuthorizerType:    authorizerTypeOrDefault(plan.AuthorizerType),
		AuthorizerConfiguration: &bactypes.AuthorizerConfigurationMemberCustomJWTAuthorizer{
			Value: bactypes.CustomJWTAuthorizerConfiguration{
				DiscoveryUrl:    aws.String(discoveryUrl),
				AllowedAudience: []string{gatewayId},
			},
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && plan.Description.ValueString() != "" {
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.ExceptionLevel.IsNull() && !plan.ExceptionLevel.IsUnknown() && plan.ExceptionLevel.ValueString() != "" {
		in.ExceptionLevel = bactypes.ExceptionLevel(plan.ExceptionLevel.ValueString())
	}
	if !plan.KmsKeyArn.IsNull() && !plan.KmsKeyArn.IsUnknown() && plan.KmsKeyArn.ValueString() != "" {
		in.KmsKeyArn = aws.String(plan.KmsKeyArn.ValueString())
	}

	out, err := conn.UpdateGateway(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating AgentCore Gateway",
			fmt.Sprintf("Could not update gateway %s: %s", gatewayId, err))
		return
	}

	r.flushOutputToState(ctx, &plan,
		gatewayId,
		state.GatewayArn.ValueString(),
		aws.ToString(out.GatewayUrl),
		string(out.Status),
		discoveryUrl,
		[]string{gatewayId},
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// -------------------------------------------------------------------
// Delete
// -------------------------------------------------------------------

func (r *BedrockAgentCoreGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BedrockAgentCoreGatewayResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := bedrockagentcorecontrol.NewFromConfig(r.config)

	_, err := conn.DeleteGateway(ctx, &bedrockagentcorecontrol.DeleteGatewayInput{
		GatewayIdentifier: aws.String(data.GatewayId.ValueString()),
	})
	if err != nil {
		var nf *bactypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting AgentCore Gateway",
			fmt.Sprintf("Could not delete gateway %s: %s", data.GatewayId.ValueString(), err))
	}
}

// -------------------------------------------------------------------
// ImportState
// -------------------------------------------------------------------

// ImportState accepts the gateway_id. Read populates the rest of the schema.
//
//	terraform import awsext_bedrockagentcore_gateway.example <gateway_id>
func (r *BedrockAgentCoreGatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be the gateway_id.",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway_id"), req.ID)...)
}

// retryUpdateGatewayUntilReady calls UpdateGateway, retrying when AWS reports
// the gateway is still in CREATING state. Backs off up to ~3 minutes.
//
// We can't poll GetGateway as a precondition because some IAM roles
// (e.g. the contractor SSO role in use) have UpdateGateway but not
// GetGateway — retrying UpdateGateway sidesteps that gap.
func retryUpdateGatewayUntilReady(
	ctx context.Context,
	conn *bedrockagentcorecontrol.Client,
	in *bedrockagentcorecontrol.UpdateGatewayInput,
) (*bedrockagentcorecontrol.UpdateGatewayOutput, error) {
	const (
		maxWait         = 3 * time.Minute
		initialWait     = 2 * time.Second
		maxBackoff      = 15 * time.Second
		creatingStateMsg = "creating state"
	)
	// Pins the matched substring so that if AWS ever changes the message text,
	// it returns an immediate error rather than silently timing out.

	deadline := time.Now().Add(maxWait)
	wait := initialWait
	for {
		out, err := conn.UpdateGateway(ctx, in)
		if err == nil {
			return out, nil
		}
		// Retry on ConflictException (gateway still provisioning) or on a
		// ValidationException whose message indicates the gateway is in CREATING state.
		if !errors.As(err, new(*bactypes.ConflictException)) {
			var ve *bactypes.ValidationException
			if !errors.As(err, &ve) {
				return nil, err
			}
			msg := ""
			if ve.Message != nil {
				msg = *ve.Message
			}
			if !strings.Contains(strings.ToLower(msg), creatingStateMsg) {
				return nil, err
			}
		}
		// ConflictException: gateway still being created — falls through to retry
		if time.Now().Add(wait).After(deadline) {
			return nil, fmt.Errorf("gateway did not leave CREATING state within %s: %w", maxWait, err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
		wait *= 2
		if wait > maxBackoff {
			wait = maxBackoff
		}
	}
}

// -------------------------------------------------------------------
// State plumbing
// -------------------------------------------------------------------

func (r *BedrockAgentCoreGatewayResource) flushOutputToState(
	ctx context.Context,
	data *BedrockAgentCoreGatewayResourceModel,
	gatewayId, gatewayArn, gatewayUrl, status, discoveryUrl string,
	audience []string,
) {
	data.GatewayId = types.StringValue(gatewayId)
	data.GatewayArn = types.StringValue(gatewayArn)
	if gatewayUrl != "" {
		data.GatewayUrl = types.StringValue(gatewayUrl)
	} else if data.GatewayUrl.IsNull() || data.GatewayUrl.IsUnknown() {
		data.GatewayUrl = types.StringValue("")
	}
	data.Status = types.StringValue(status)
	data.CustomJwtAuthorizer = &BedrockAgentCoreCustomJWTAuthorizerModel{
		DiscoveryUrl: types.StringValue(discoveryUrl),
	}
	if len(audience) > 0 {
		lv, diags := types.ListValueFrom(ctx, types.StringType, audience)
		if !diags.HasError() {
			data.AllowedAudience = lv
		} else {
			data.AllowedAudience = types.ListNull(types.StringType)
		}
	} else {
		data.AllowedAudience = types.ListNull(types.StringType)
	}
}

// findGatewayIdByName pages through ListGateways until it finds a gateway
// whose name matches, then returns its gateway_id.
func findGatewayIdByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (string, error) {
	var nextToken *string
	for {
		out, err := conn.ListGateways(ctx, &bedrockagentcorecontrol.ListGatewaysInput{
			NextToken: nextToken,
		})
		if err != nil {
			return "", err
		}
		for _, gw := range out.Items {
			if aws.ToString(gw.Name) == name {
				return aws.ToString(gw.GatewayId), nil
			}
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return "", fmt.Errorf("no gateway with name %q found", name)
}
