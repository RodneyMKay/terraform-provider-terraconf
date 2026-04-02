package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &TerraconfProvider{}
var _ provider.ProviderWithFunctions = &TerraconfProvider{}

type TerraconfProvider struct {
	version string
}

func (p *TerraconfProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "terraconf"
	resp.Version = p.version
}

func (p *TerraconfProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (p *TerraconfProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *TerraconfProvider) Resources(ctx context.Context) []func() resource.Resource {
	// This provider doesn't have any resources
	return []func() resource.Resource{}
}

func (p *TerraconfProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewYamlDataSource,
	}
}

func (p *TerraconfProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewSanatizeFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TerraconfProvider{
			version: version,
		}
	}
}
