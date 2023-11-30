package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	devops_resource "github.com/liatrio/devops-bootcamp/examples/ch7/devops-resources"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &engineersResource{}
	_ resource.ResourceWithConfigure = &engineersResource{}
)

// NewEngineersResource is a helper function to simplify the provider implementation.
func NewEngineersResource() resource.Resource {
	return &engineersResource{}
}

// EngineersResource is the resource implementation.
type engineersResource struct {
	client *http.Client
}

// EngineersResourceModel describes the resource data model.
type engineersResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Email       types.String `tfsdk:"email"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *engineersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_engineers"
}

// Configure adds the provider configured client to the resource.
func (r *engineersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *engineersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"email": schema.StringAttribute{
				Required: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *engineersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan *engineersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	engineerItem := map[string]string{"name": plan.Name.ValueString(), "email": plan.Email.ValueString()}

	jsonBody, err := json.Marshal(engineerItem)
	//url := r.Host + "/engineers"

	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080/engineers", bytes.NewBuffer(jsonBody))
	response, err := r.client.Do(request)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating engineer",
			"Could not create engineer: "+err.Error(),
		)
		return
	}

	// Return error if the HTTP status code is not 201 OK
	if response.StatusCode != 201 {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status,
		)
		return
	}

	var engineer devops_resource.Engineer
	err = json.NewDecoder(response.Body).Decode(&engineer)
	if err != nil {
		// do something
	}
	//Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(engineer.Id)
	plan.Name = types.StringValue(engineer.Name)
	plan.Email = types.StringValue(engineer.Email)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *engineersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state engineersResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("http://localhost:8080/engineers/name/%s", state.Name.ValueString())

	request, err := http.NewRequest(http.MethodGet, url, nil)
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading engineers",
			"Could not read engineers: "+err.Error(),
		)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)

		return
	}

	var engineer devops_resource.Engineer
	err = json.NewDecoder(response.Body).Decode(&engineer)

	// Overwrite items with refreshed state
	state.ID = types.StringValue(engineer.Id)
	state.Name = types.StringValue(engineer.Name)
	state.Email = types.StringValue(engineer.Email)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *engineersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan *engineersResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("http://localhost:8080/engineers/%s", plan.ID.ValueString())

	engineerItem := map[string]string{"name": plan.Name.ValueString(), "email": plan.Email.ValueString()}
	jsonBody, err := json.Marshal(engineerItem)

	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating engineers",
			"Could not update engineers: "+err.Error(),
		)
		return
	}

	defer response.Body.Close()

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status,
		)
		return
	}

	var engineer devops_resource.Engineer
	err = json.NewDecoder(response.Body).Decode(&engineer)
	if err != nil {
		// do something
	}
	//Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(engineer.Id)
	plan.Name = types.StringValue(engineer.Name)
	plan.Email = types.StringValue(engineer.Email)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *engineersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from plan
	var plan *engineersResourceModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("http://localhost:8080/engineers/%s", plan.ID.ValueString())

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading engineers",
			"Could not read engineers: "+err.Error(),
		)
		return
	}

	defer response.Body.Close()

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status,
		)
		return
	}

	resp.State.RemoveResource(ctx)

	return
}
