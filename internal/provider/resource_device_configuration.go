package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/Nathan-Loisel/terraform-provider-webex/internal/client"
)

var (
	_ resource.Resource                = &DeviceConfigurationResource{}
	_ resource.ResourceWithImportState = &DeviceConfigurationResource{}
)

type DeviceConfigurationResource struct {
	client *client.Client
}

type DeviceConfigurationResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DeviceID       types.String `tfsdk:"device_id"`
	Configurations types.Map    `tfsdk:"configurations"`
}

func NewDeviceConfigurationResource() resource.Resource {
	return &DeviceConfigurationResource{}
}

func (r *DeviceConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_configuration"
}

func (r *DeviceConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages device configurations (xAPI settings) for a Webex device. Configurations are key-value pairs that control device behavior.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (same as device_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Required:    true,
				Description: "The device to configure.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configurations": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Map of configuration keys to values. Keys use dot-separated paths (e.g., 'Standby.Delay', 'Audio.Ultrasound.MaxVolume'). Values are always strings. Removing a key from this map reverts it to the device default.",
			},
		},
	}
}

func (r *DeviceConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *DeviceConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeviceConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := plan.DeviceID.ValueString()

	var configs map[string]string
	resp.Diagnostics.Append(plan.Configurations.ElementsAs(ctx, &configs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(configs) > 0 {
		supported, err := r.supportedKeys(ctx, deviceID)
		if err != nil {
			resp.Diagnostics.AddError("Error reading device capabilities", err.Error())
			return
		}

		ops := make([]client.ConfigurationPatchOp, 0, len(configs))
		for key, value := range configs {
			if _, ok := supported[key]; ok {
				ops = append(ops, client.ConfigurationPatchOp{
					Op:    "replace",
					Path:  key + "/sources/configured/value",
					Value: typedValue(value),
				})
			}
		}

		if len(ops) > 0 {
			_, err := r.client.UpdateDeviceConfigurations(ctx, deviceID, ops)
			if err != nil {
				resp.Diagnostics.AddError("Error setting device configurations", err.Error())
				return
			}
		}
	}

	plan.ID = types.StringValue(deviceID)
	r.refreshConfigState(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeviceConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshConfigState(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeviceConfigurationResourceModel
	var state DeviceConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := plan.DeviceID.ValueString()

	var planConfigs map[string]string
	var stateConfigs map[string]string
	resp.Diagnostics.Append(plan.Configurations.ElementsAs(ctx, &planConfigs, false)...)
	resp.Diagnostics.Append(state.Configurations.ElementsAs(ctx, &stateConfigs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supported, err := r.supportedKeys(ctx, deviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading device capabilities", err.Error())
		return
	}

	ops := make([]client.ConfigurationPatchOp, 0)

	for key, value := range planConfigs {
		if _, ok := supported[key]; !ok {
			continue
		}
		if oldVal, exists := stateConfigs[key]; !exists || oldVal != value {
			ops = append(ops, client.ConfigurationPatchOp{
				Op:    "replace",
				Path:  key + "/sources/configured/value",
				Value: typedValue(value),
			})
		}
	}

	for key := range stateConfigs {
		if _, exists := planConfigs[key]; !exists {
			if _, ok := supported[key]; ok {
				ops = append(ops, client.ConfigurationPatchOp{
					Op:   "remove",
					Path: key + "/sources/configured/value",
				})
			}
		}
	}

	if len(ops) > 0 {
		_, err := r.client.UpdateDeviceConfigurations(ctx, deviceID, ops)
		if err != nil {
			resp.Diagnostics.AddError("Error updating device configurations", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(deviceID)
	r.refreshConfigState(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeviceConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := state.DeviceID.ValueString()

	var configs map[string]string
	resp.Diagnostics.Append(state.Configurations.ElementsAs(ctx, &configs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(configs) > 0 {
		ops := make([]client.ConfigurationPatchOp, 0, len(configs))
		for key := range configs {
			ops = append(ops, client.ConfigurationPatchOp{
				Op:   "remove",
				Path: key + "/sources/configured/value",
			})
		}

		_, err := r.client.UpdateDeviceConfigurations(ctx, deviceID, ops)
		if err != nil && !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Error reverting device configurations to defaults", err.Error())
		}
	}
}

func (r *DeviceConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), req.ID)...)
}

func (r *DeviceConfigurationResource) refreshConfigState(ctx context.Context, model *DeviceConfigurationResourceModel, diagnostics *diag.Diagnostics) {
	deviceID := model.DeviceID.ValueString()

	// Get the keys we're managing from the current plan/state
	var managedKeys map[string]string
	diagnostics.Append(model.Configurations.ElementsAs(ctx, &managedKeys, false)...)
	if diagnostics.HasError() {
		return
	}

	configResp, err := r.client.GetDeviceConfigurations(ctx, deviceID, "")
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		diagnostics.AddError("Error reading device configurations", err.Error())
		return
	}

	result := make(map[string]string)
	if len(managedKeys) == 0 {
		// Import: discover all keys that have a configured source value
		for key, cfg := range configResp.Items {
			if src, ok := cfg.Sources["configured"]; ok {
				result[key] = fmt.Sprintf("%v", src.Value)
			}
		}
	} else {
		// Normal read: only track keys we're managing — read from the
		// "configured" source so we detect drift on the exact value Terraform wrote.
		for key, plannedValue := range managedKeys {
			if cfg, exists := configResp.Items[key]; exists {
				if src, ok := cfg.Sources["configured"]; ok {
					result[key] = fmt.Sprintf("%v", src.Value)
				} else {
					result[key] = fmt.Sprintf("%v", cfg.Value)
				}
			} else {
				result[key] = plannedValue
			}
		}
	}

	configMap, diags := types.MapValueFrom(ctx, types.StringType, result)
	diagnostics.Append(diags...)
	model.Configurations = configMap
}

func (r *DeviceConfigurationResource) supportedKeys(ctx context.Context, deviceID string) (map[string]bool, error) {
	configResp, err := r.client.GetDeviceConfigurations(ctx, deviceID, "")
	if err != nil {
		return nil, err
	}
	keys := make(map[string]bool, len(configResp.Items))
	for key := range configResp.Items {
		keys[key] = true
	}
	return keys, nil
}

func typedValue(s string) interface{} {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return s
}
