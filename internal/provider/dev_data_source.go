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
	_ datasource.DataSource              = &devDataSource{}
	_ datasource.DataSourceWithConfigure = &devDataSource{}
)

func NewDevDataSource() datasource.DataSource {
	return &devDataSource{}
}

// devDataSource defines the data source implementation.
type devDataSource struct {
	client *http.Client
}

// devDataSourceModel describes the data source data model.
type devDataSourceModel struct {
	devs []devModel `tfsdk:"devs"`
}

// devModel maps dev schema data.
type devModel struct {
	ID        types.String     `tfsdk:"id"`
	Name      types.String     `tfsdk:"name"`
	Engineers []engineersModel `tfsdk:"engineers"`
}

type devModelLocal struct {
	ID    string `tfsdk:"id"`
	Name  string `tfsdk:"name"`
	Email string `tfsdk:"email"`
}

func (d *devDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dev"
}

func (d *devDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"devs": schema.ListNestedAttribute{
				MarkdownDescription: "Dev attribute",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
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
										Required: true,
									},
								},
							},
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *devDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *devDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state devDataSourceModel

	request, err := http.NewRequest(http.MethodGet, "http://localhost:8080/dev", nil)
	if err != nil {
		return
	}

	response, err := d.client.Do(request)
	if err != nil {
		return
	}

	defer response.Body.Close() // Close the response body when done

	var devs []devops_resource.Dev
	err = json.NewDecoder(response.Body).Decode(&devs)
	if err != nil {
		return
	}

	// Map response body to model
	for _, dev := range devs {
		temp := devModel{
			ID:   types.StringValue(dev.Id),
			Name: types.StringValue(dev.Name),
		}
		for _, engineer := range dev.Engineers {
			tempEngineer := engineersModel{
				ID:    types.StringValue(engineer.Id),
				Name:  types.StringValue(engineer.Name),
				Email: types.StringValue(engineer.Email),
			}
			temp.Engineers = append(temp.Engineers, tempEngineer)
		}
		state.devs = append(state.devs, temp)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
