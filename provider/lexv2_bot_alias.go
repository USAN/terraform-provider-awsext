package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	lextypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &LexV2BotAliasResource{}
var _ resource.ResourceWithImportState = &LexV2BotAliasResource{}

func NewLexV2BotAliasResource() resource.Resource {
	return &LexV2BotAliasResource{}
}

type LexV2BotAliasResource struct {
	config aws.Config
}

// -------------------------------------------------------------------
// Model
// -------------------------------------------------------------------

type LexV2BotAliasResourceModel struct {
	BotID                  types.String `tfsdk:"bot_id"`
	Name                   types.String `tfsdk:"name"`
	BotVersion             types.String `tfsdk:"bot_version"`
	LocaleID               types.String `tfsdk:"locale_id"`
	LocaleEnabled          types.Bool   `tfsdk:"locale_enabled"`
	TextLogEnabled         types.Bool   `tfsdk:"text_log_enabled"`
	TextLogCWLogGroupARN   types.String `tfsdk:"text_log_cw_log_group_arn"`
	TextLogPrefix          types.String `tfsdk:"text_log_prefix"`
	Tags                   types.Map    `tfsdk:"tags"`
	BotAliasID             types.String `tfsdk:"bot_alias_id"`
	Arn                    types.String `tfsdk:"arn"`
	AdoptOnExists          types.Bool   `tfsdk:"adopt_on_exists"`
}

// -------------------------------------------------------------------
// Metadata / Schema
// -------------------------------------------------------------------

func (r *LexV2BotAliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lexv2_bot_alias"
}

func (r *LexV2BotAliasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and manages an Amazon Lex V2 bot alias " +
			"(lexmodelsv2:CreateBotAlias / UpdateBotAlias / DescribeBotAlias / DeleteBotAlias). " +
			"Fills the gap left by the official AWS provider which does not include aws_lexv2models_bot_alias.",

		Attributes: map[string]schema.Attribute{
			"bot_id": schema.StringAttribute{
				Required:    true,
				Description: "The identifier of the bot to create the alias for. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The alias name. Must be unique for the bot. Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bot_version": schema.StringAttribute{
				Optional:    true,
				Description: "The bot version the alias points to (e.g. '1', '2'). Omit or set to empty to leave the alias without a pinned version (DRAFT traffic).",
			},
			"locale_id": schema.StringAttribute{
				Required:    true,
				Description: "The locale ID to enable on this alias (e.g. 'en_US'). Forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locale_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the locale is enabled on this alias. Defaults to true.",
			},
			"text_log_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether CloudWatch text logging is enabled. Requires text_log_cw_log_group_arn and text_log_prefix when true.",
			},
			"text_log_cw_log_group_arn": schema.StringAttribute{
				Optional:    true,
				Description: "CloudWatch log group ARN for text logs. Required when text_log_enabled is true.",
			},
			"text_log_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "Log stream prefix for CloudWatch text logs. Required when text_log_enabled is true.",
			},
			"tags": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags to assign to the bot alias. Updatable in place.",
			},
			"bot_alias_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service-assigned identifier of the bot alias.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"arn": schema.StringAttribute{
				Computed:    true,
				Description: "The Amazon Resource Name (ARN) of the bot alias.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"adopt_on_exists": schema.BoolAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "When true, if a bot alias with the given name already exists for the bot, adopt it into state instead of attempting to create it. Not stored in state.",
			},
		},
	}
}

// -------------------------------------------------------------------
// Configure
// -------------------------------------------------------------------

func (r *LexV2BotAliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LexV2BotAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LexV2BotAliasResourceModel
	var adoptOnExists types.Bool
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("adopt_on_exists"), &adoptOnExists)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	// If adopt_on_exists is set, check for an existing alias with this name and adopt it.
	if !adoptOnExists.IsNull() && !adoptOnExists.IsUnknown() && adoptOnExists.ValueBool() {
		aliasID, found, findErr := r.findAliasByName(ctx, client, data.BotID.ValueString(), data.Name.ValueString())
		if findErr != nil {
			resp.Diagnostics.AddError("Error searching for existing bot alias", findErr.Error())
			return
		}
		if found {
			tflog.Info(ctx, "adopt_on_exists: found existing alias, adopting into state",
				map[string]any{"bot_alias_id": aliasID})
			data.BotAliasID = types.StringValue(aliasID)
			aliasArn, arnErr := r.buildAliasArn(ctx, data.BotID.ValueString(), aliasID)
			if arnErr != nil {
				resp.Diagnostics.AddError("Error constructing bot alias ARN", arnErr.Error())
				return
			}
			data.Arn = types.StringValue(aliasArn)
			// Sync plan tags to the adopted alias. Read existing tags only to compute
			// the diff; final state uses plan.Tags directly to avoid AWS normalisation
			// differences (e.g. "True"→"true") and eventual-consistency read-back issues.
			existingTags, tagsErr := readLexV2Tags(ctx, client, aliasArn)
			if tagsErr != nil {
				resp.Diagnostics.AddError("Error reading bot alias tags", tagsErr.Error())
				return
			}
			if err := updateLexV2Tags(ctx, client, aliasArn, existingTags, data.Tags); err != nil {
				resp.Diagnostics.AddError("Error syncing bot alias tags on adopt", err.Error())
				return
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	in := &lexmodelsv2.CreateBotAliasInput{
		BotId:        aws.String(data.BotID.ValueString()),
		BotAliasName: aws.String(data.Name.ValueString()),
		BotAliasLocaleSettings: map[string]lextypes.BotAliasLocaleSettings{
			data.LocaleID.ValueString(): {
				Enabled: data.LocaleEnabled.ValueBool(),
			},
		},
	}

	if !data.BotVersion.IsNull() && !data.BotVersion.IsUnknown() && data.BotVersion.ValueString() != "" {
		in.BotVersion = aws.String(data.BotVersion.ValueString())
	}

	if data.TextLogEnabled.ValueBool() {
		in.ConversationLogSettings = r.buildConversationLogSettings(data)
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tagMap map[string]string
		if diags := data.Tags.ElementsAs(ctx, &tagMap, false); !diags.HasError() && len(tagMap) > 0 {
			in.Tags = tagMap
		}
	}

	out, err := client.CreateBotAlias(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Lex V2 bot alias", err.Error())
		return
	}

	aliasID := aws.ToString(out.BotAliasId)
	data.BotAliasID = types.StringValue(aliasID)

	aliasArn, arnErr := r.buildAliasArn(ctx, data.BotID.ValueString(), aliasID)
	if arnErr != nil {
		resp.Diagnostics.AddError("Error constructing bot alias ARN", arnErr.Error())
		return
	}
	data.Arn = types.StringValue(aliasArn)

	if err := r.pollUntilAvailable(ctx, client, data.BotID.ValueString(), aliasID); err != nil {
		resp.Diagnostics.AddError("Error waiting for bot alias to become available", err.Error())
		return
	}

	tags, tagsErr := readLexV2Tags(ctx, client, aliasArn)
	if tagsErr != nil {
		resp.Diagnostics.AddError("Error reading bot alias tags", tagsErr.Error())
		return
	}
	data.Tags = tags

	tflog.Trace(ctx, "created lexv2 bot alias", map[string]any{"bot_alias_id": aliasID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Read
// -------------------------------------------------------------------

func (r *LexV2BotAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LexV2BotAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	out, err := client.DescribeBotAlias(ctx, &lexmodelsv2.DescribeBotAliasInput{
		BotId:      aws.String(data.BotID.ValueString()),
		BotAliasId: aws.String(data.BotAliasID.ValueString()),
	})
	if err != nil {
		var nf *lextypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			tflog.Warn(ctx, "Lex V2 bot alias not found, removing from state",
				map[string]any{"bot_alias_id": data.BotAliasID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Lex V2 bot alias", err.Error())
		return
	}

	data.Name = types.StringValue(aws.ToString(out.BotAliasName))
	if out.BotVersion != nil {
		data.BotVersion = types.StringValue(aws.ToString(out.BotVersion))
	} else {
		data.BotVersion = types.StringNull()
	}

	if localeSettings, ok := out.BotAliasLocaleSettings[data.LocaleID.ValueString()]; ok {
		data.LocaleEnabled = types.BoolValue(localeSettings.Enabled)
	}

	if out.ConversationLogSettings != nil && len(out.ConversationLogSettings.TextLogSettings) > 0 {
		tls := out.ConversationLogSettings.TextLogSettings[0]
		data.TextLogEnabled = types.BoolValue(tls.Enabled)
		if tls.Destination.CloudWatch != nil {
			data.TextLogCWLogGroupARN = types.StringValue(aws.ToString(tls.Destination.CloudWatch.CloudWatchLogGroupArn))
			data.TextLogPrefix = types.StringValue(aws.ToString(tls.Destination.CloudWatch.LogPrefix))
		}
	} else {
		data.TextLogEnabled = types.BoolValue(false)
	}

	if data.Arn.IsNull() || data.Arn.IsUnknown() || data.Arn.ValueString() == "" {
		aliasArn, arnErr := r.buildAliasArn(ctx, data.BotID.ValueString(), data.BotAliasID.ValueString())
		if arnErr == nil {
			data.Arn = types.StringValue(aliasArn)
		}
	}

	tags, tagsErr := readLexV2Tags(ctx, client, data.Arn.ValueString())
	if tagsErr != nil {
		resp.Diagnostics.AddError("Error reading bot alias tags", tagsErr.Error())
		return
	}
	data.Tags = tags

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -------------------------------------------------------------------
// Update
// -------------------------------------------------------------------

func (r *LexV2BotAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LexV2BotAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	in := &lexmodelsv2.UpdateBotAliasInput{
		BotId:        aws.String(state.BotID.ValueString()),
		BotAliasId:   aws.String(state.BotAliasID.ValueString()),
		BotAliasName: aws.String(state.Name.ValueString()),
		BotAliasLocaleSettings: map[string]lextypes.BotAliasLocaleSettings{
			plan.LocaleID.ValueString(): {
				Enabled: plan.LocaleEnabled.ValueBool(),
			},
		},
	}

	if !plan.BotVersion.IsNull() && !plan.BotVersion.IsUnknown() && plan.BotVersion.ValueString() != "" {
		in.BotVersion = aws.String(plan.BotVersion.ValueString())
	}

	if plan.TextLogEnabled.ValueBool() {
		in.ConversationLogSettings = r.buildConversationLogSettings(plan)
	}

	if _, err := client.UpdateBotAlias(ctx, in); err != nil {
		resp.Diagnostics.AddError("Error updating Lex V2 bot alias", err.Error())
		return
	}

	if err := updateLexV2Tags(ctx, client, state.Arn.ValueString(), state.Tags, plan.Tags); err != nil {
		resp.Diagnostics.AddError("Error updating bot alias tags", err.Error())
		return
	}

	plan.BotAliasID = state.BotAliasID
	plan.Arn = state.Arn
	// Use plan.Tags directly — avoids AWS tag-value normalisation differences
	// (e.g. "True"→"true") and eventual-consistency gaps from ListTagsForResource.
	// The updateLexV2Tags call above already made the resource match plan.Tags.

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// -------------------------------------------------------------------
// Delete
// -------------------------------------------------------------------

func (r *LexV2BotAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LexV2BotAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := lexmodelsv2.NewFromConfig(r.config)

	_, err := client.DeleteBotAlias(ctx, &lexmodelsv2.DeleteBotAliasInput{
		BotId:                  aws.String(data.BotID.ValueString()),
		BotAliasId:             aws.String(data.BotAliasID.ValueString()),
		SkipResourceInUseCheck: false,
	})
	if err != nil {
		var nf *lextypes.ResourceNotFoundException
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting Lex V2 bot alias", err.Error())
	}
}

// -------------------------------------------------------------------
// ImportState
// -------------------------------------------------------------------

// ImportState accepts the import ID in the format <bot_id>/<bot_alias_id>.
func (r *LexV2BotAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Minimal import: caller must provide bot_id and bot_alias_id separated by '/'.
	// The subsequent Read will fill remaining computed attributes.
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be <bot_id>/<bot_alias_id>, got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bot_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bot_alias_id"), parts[1])...)
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func (r *LexV2BotAliasResource) buildAliasArn(ctx context.Context, botID, aliasID string) (string, error) {
	stsClient := sts.NewFromConfig(r.config)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("GetCallerIdentity: %w", err)
	}
	return fmt.Sprintf("arn:aws:lex:%s:%s:bot-alias/%s/%s",
		r.config.Region, aws.ToString(identity.Account), botID, aliasID), nil
}

func (r *LexV2BotAliasResource) buildConversationLogSettings(data LexV2BotAliasResourceModel) *lextypes.ConversationLogSettings {
	if data.TextLogCWLogGroupARN.IsNull() || data.TextLogCWLogGroupARN.ValueString() == "" {
		return nil
	}
	return &lextypes.ConversationLogSettings{
		TextLogSettings: []lextypes.TextLogSetting{
			{
				Enabled: data.TextLogEnabled.ValueBool(),
				Destination: &lextypes.TextLogDestination{
					CloudWatch: &lextypes.CloudWatchLogGroupLogDestination{
						CloudWatchLogGroupArn: aws.String(data.TextLogCWLogGroupARN.ValueString()),
						LogPrefix:             aws.String(data.TextLogPrefix.ValueString()),
					},
				},
			},
		},
	}
}

func (r *LexV2BotAliasResource) findAliasByName(ctx context.Context, client *lexmodelsv2.Client, botID, name string) (aliasID string, found bool, err error) {
	var nextToken *string
	for {
		out, listErr := client.ListBotAliases(ctx, &lexmodelsv2.ListBotAliasesInput{
			BotId:     aws.String(botID),
			NextToken: nextToken,
		})
		if listErr != nil {
			return "", false, fmt.Errorf("ListBotAliases: %w", listErr)
		}
		for _, summary := range out.BotAliasSummaries {
			if aws.ToString(summary.BotAliasName) == name {
				return aws.ToString(summary.BotAliasId), true, nil
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return "", false, nil
}

func (r *LexV2BotAliasResource) pollUntilAvailable(ctx context.Context, client *lexmodelsv2.Client, botID, aliasID string) error {
	const maxIterations = 30
	const sleep = 5 * time.Second
	for i := 0; i < maxIterations; i++ {
		time.Sleep(sleep)
		out, err := client.DescribeBotAlias(ctx, &lexmodelsv2.DescribeBotAliasInput{
			BotId:      aws.String(botID),
			BotAliasId: aws.String(aliasID),
		})
		if err != nil {
			return fmt.Errorf("DescribeBotAlias: %w", err)
		}
		switch out.BotAliasStatus {
		case lextypes.BotAliasStatusAvailable:
			return nil
		case lextypes.BotAliasStatusFailed:
			return fmt.Errorf("bot alias %s reached Failed status", aliasID)
		}
	}
	return fmt.Errorf("timed out waiting for bot alias %s to become Available", aliasID)
}
