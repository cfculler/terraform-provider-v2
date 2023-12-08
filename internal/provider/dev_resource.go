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
	_ resource.Resource              = &devResource{}
	_ resource.ResourceWithConfigure = &devResource{}
)

// NewdevResource is a helper function to simplify the provider implementation.
func NewDevResource() resource.Resource {
	return &devResource{}
}

// devResource is the resource implementation.
type devResource struct {
	client *http.Client
}

// devResourceModel describes the resource data model.
type devResourceModel struct {
	ID          types.String     `tfsdk:"id"`
	Name        types.String     `tfsdk:"name"`
	Engineers   []engineersModel `tfsdk:"engineers"`
	LastUpdated types.String     `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *devResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dev"
}

// Configure adds the provider configured client to the resource.
func (r *devResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *devResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"engineers": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"email": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *devResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan *devResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	devItem := map[string]string{"name": plan.Name.ValueString()}

	jsonBody, err := json.Marshal(devItem)
	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080/dev", bytes.NewBuffer(jsonBody))

	response, err := r.client.Do(request)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating dev",
			"Could not create dev: "+err.Error(),
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

	var dev devops_resource.Dev
	err = json.NewDecoder(response.Body).Decode(&dev)
	if err != nil {
		// do something
	}
	//Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(dev.Id)
	plan.Name = types.StringValue(dev.Name)
	url := fmt.Sprintf("http://localhost:8080/dev/%s", plan.ID.ValueString())

	var dev2 devops_resource.Dev
	for _, engineer := range plan.Engineers {
		tempUrl := fmt.Sprintf("http://localhost:8080/engineers/name/%s", engineer.Name.ValueString())

		request, err := http.NewRequest(http.MethodGet, tempUrl, nil)
		response, err := r.client.Do(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading engineer",
				"Could not read engineer: "+err.Error(),
			)
			return
		}

		if response.StatusCode == http.StatusNotFound {
			//resp.State.RemoveResource(ctx)
			response.Body.Close()
			continue
		}

		var engineer devops_resource.Engineer
		err = json.NewDecoder(response.Body).Decode(&engineer)
		response.Body.Close()

		// Overwrite items with refreshed state
		temp := engineer.Id
		engineerItem := map[string]string{"id": temp}
		jsonBody, err = json.Marshal(engineerItem)
		request, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
		response, err = r.client.Do(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding engineer to dev",
				"Could not create dev: "+err.Error(),
			)
			continue
		}

		// Return error if the HTTP status code is not 200 OK
		if response.StatusCode != 200 {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while attempting to create the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"HTTP Status: "+response.Status,
			)
			continue
		}
		err = json.NewDecoder(response.Body).Decode(&dev2)
		response.Body.Close()
	}

	//Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(dev2.Id)
	//plan.Name = types.StringValue(dev2.Name)
	plan.Engineers = plan.Engineers[:0]
	for _, engineer := range dev2.Engineers {
		temp := engineersModel{
			ID:    types.StringValue(engineer.Id),
			Name:  types.StringValue(engineer.Name),
			Email: types.StringValue(engineer.Email),
		}
		plan.Engineers = append(plan.Engineers, temp)
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *devResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state devResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("http://localhost:8080/dev/name/%s", state.Name.ValueString())

	request, err := http.NewRequest(http.MethodGet, url, nil)
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading dev",
			"Could not read dev: "+err.Error(),
		)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)

		return
	}

	var dev devops_resource.Dev
	err = json.NewDecoder(response.Body).Decode(&dev)

	// Overwrite items with refreshed state
	state.ID = types.StringValue(dev.Id)
	state.Name = types.StringValue(dev.Name)
	state.Engineers = state.Engineers[:0]
	for _, engineer := range dev.Engineers {
		temp := engineersModel{
			ID:    types.StringValue(engineer.Id),
			Name:  types.StringValue(engineer.Name),
			Email: types.StringValue(engineer.Email),
		}
		state.Engineers = append(state.Engineers, temp)
	}
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *devResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan *devResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	devItem := map[string]string{"name": plan.Name.ValueString()}
	url := fmt.Sprintf("http://localhost:8080/dev/%s", plan.ID.ValueString())

	jsonBody, err := json.Marshal(devItem)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))

	response, err := r.client.Do(request)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating dev",
			"Could not create dev: "+err.Error(),
		)
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode != 200 {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status,
		)
		return
	}

	var dev devops_resource.Dev
	err = json.NewDecoder(response.Body).Decode(&dev)
	if err != nil {
		// do something
	}
	//Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(dev.Id)
	plan.Name = types.StringValue(dev.Name)

	var dev2 devops_resource.Dev
	for _, engineer := range plan.Engineers {
		tempUrl := fmt.Sprintf("http://localhost:8080/engineers/name/%s", engineer.Name.ValueString())

		request, err := http.NewRequest(http.MethodGet, tempUrl, nil)
		response, err := r.client.Do(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading engineer",
				"Could not read engineer: "+err.Error(),
			)
			return
		}

		if response.StatusCode == http.StatusNotFound {
			//resp.State.RemoveResource(ctx)
			response.Body.Close()
			continue
		}

		var engineer devops_resource.Engineer
		err = json.NewDecoder(response.Body).Decode(&engineer)
		response.Body.Close()

		// Overwrite items with refreshed state
		temp := engineer.Id
		engineerItem := map[string]string{"id": temp}
		jsonBody, err = json.Marshal(engineerItem)

		request, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
		response, err = r.client.Do(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding engineer to dev",
				"Could not create dev: "+err.Error(),
			)
			continue
		}

		// Return error if the HTTP status code is not 200 OK
		if response.StatusCode != 200 {
			resp.Diagnostics.AddError(
				"Unable to Create Resource",
				"An unexpected error occurred while attempting to create the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"HTTP Status: "+response.Status,
			)
			continue
		}
		err = json.NewDecoder(response.Body).Decode(&dev2)
		response.Body.Close()
	}

	//Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(dev2.Id)
	plan.Name = types.StringValue(dev2.Name)
	plan.Engineers = plan.Engineers[:0]
	for _, engineer := range dev2.Engineers {
		temp := engineersModel{
			ID:    types.StringValue(engineer.Id),
			Name:  types.StringValue(engineer.Name),
			Email: types.StringValue(engineer.Email),
		}
		plan.Engineers = append(plan.Engineers, temp)
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *devResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from plan
	var plan *devResourceModel
	diags := req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("http://localhost:8080/dev/%s", plan.ID.ValueString())

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading dev",
			"Could not read dev: "+err.Error(),
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
