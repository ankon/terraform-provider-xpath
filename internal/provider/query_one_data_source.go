// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/antchfx/xmlquery"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type nodeModel struct {
	Type         string           `tfsdk:"type"`
	Name         string           `tfsdk:"name"`
	Prefix       string           `tfsdk:"prefix"`
	NamespaceURI string           `tfsdk:"namespace_uri"`
	Attributes   []attributeModel `tfsdk:"attributes"`
	Data         string           `tfsdk:"data"`
}

type attributeModel struct {
	Name         string `tfsdk:"name"`
	Prefix       string `tfsdk:"prefix"`
	NamespaceURI string `tfsdk:"namespace_uri"`
	Data         string `tfsdk:"data"`
}

func modelFromAttributes(attr []xmlquery.Attr) (attributes []attributeModel) {
	for _, a := range attr {
		attributes = append(attributes, attributeModel{
			Name:         a.Name.Local,
			Prefix:       a.Name.Space,
			NamespaceURI: a.NamespaceURI,
			Data:         a.Value,
		})
	}
	return
}

func modelFromNode(n *xmlquery.Node) nodeModel {
	attributes := modelFromAttributes(n.Attr)

	result := nodeModel{
		Type:         nodeTypeToName(n.Type),
		Data:         n.Data,
		Prefix:       n.Prefix,
		NamespaceURI: n.NamespaceURI,
		Attributes:   attributes,
	}
	if n.Type == xmlquery.ElementNode {
		// "Data" contains the name of the node, so shuffle this around for easier usage
		result.Name = result.Data
		result.Data = ""
	}
	return result
}

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSourceWithConfigValidators = &QueryOneDataSource{}

func NewQueryOneDataSource() datasource.DataSource {
	return &QueryOneDataSource{}
}

// QueryOneDataSource defines the data source implementation.
type QueryOneDataSource struct{}

func (d *QueryOneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_query_one"
}

func (d *QueryOneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"expression": schema.StringAttribute{
				MarkdownDescription: "An XPath expression",
				Required:            true,
			},
			"content": schema.StringAttribute{
				Optional: true,
			},
			"content_url": schema.StringAttribute{
				Optional: true,
			},
			"namespace_bindings": schema.MapAttribute{
				ElementType: basetypes.StringType{},
				Optional:    true,
				Validators: []validator.Map{
					// Validate this map must contain only non-empty strings.
					mapvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"result": schema.DynamicAttribute{
				Computed: true,
			},
		},
	}
}

func (d *QueryOneDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("content"),
			path.MatchRoot("content_url"),
		),
	}
}

func (d *QueryOneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
}

func (d *QueryOneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data XPathDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nodes, err := xpathQueryHelper(ctx, data)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Failed to process query", err.Error()))
		return
	}

	if len(nodes) != 1 {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Unexpected result", fmt.Sprintf("found %d nodes, expected 1", len(nodes))))
		return
	}

	// Translate the node into the result
	n := nodes[0]
	ov, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"type":          basetypes.StringType{},
		"name":          basetypes.StringType{},
		"prefix":        basetypes.StringType{},
		"namespace_uri": basetypes.StringType{},
		"data":          basetypes.StringType{},
		"attributes": basetypes.ListType{ElemType: basetypes.ObjectType{AttrTypes: map[string]attr.Type{
			"name":          basetypes.StringType{},
			"prefix":        basetypes.StringType{},
			"namespace_uri": basetypes.StringType{},
			"value":         basetypes.StringType{},
		}}},
	}, modelFromNode(n))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Result = types.DynamicValue(ov)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
