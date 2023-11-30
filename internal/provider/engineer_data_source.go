// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	devops_resource "github.com/liatrio/devops-bootcamp/examples/ch7/devops-resources"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &engineersDataSource{}
	_ datasource.DataSourceWithConfigure = &engineersDataSource{}
)

func NewEngineersDataSource() datasource.DataSource {
	return &engineersDataSource{}
}

// engineersDataSource defines the data source implementation.
type engineersDataSource struct {
	client *http.Client
}

// engineersDataSourceModel describes the data source data model.
// type engineersDataSourceModel struct {
// 	Engineers []engineersModel `tfsdk:"engineers"`
// }

// engineersModel maps engineers schema data.
type engineersModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

type engineersModelLocal struct {
	ID    string `tfsdk:"id"`
	Name  string `tfsdk:"name"`
	Email string `tfsdk:"email"`
}

func (d *engineersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_engineers"
}

func (d *engineersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"engineers": schema.ListNestedAttribute{
				MarkdownDescription: "Engineer attribute",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"email": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (d *engineersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *engineersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state engineersModel

	request, err := http.NewRequest(http.MethodGet, "http://localhost:8080/engineers", nil)
	if err != nil {
		return
	}

	response, err := d.client.Do(request)
	if err != nil {
		return
	}

	defer response.Body.Close() // Close the response body when done

	var engineer devops_resource.Engineer
	err = json.NewDecoder(response.Body).Decode(&engineer)
	if err != nil {
		return
	}

	// Map response body to model
	state.ID = types.StringValue(engineer.Id)
	state.Name = types.StringValue(engineer.Name)
	state.Email = types.StringValue(engineer.Email)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
