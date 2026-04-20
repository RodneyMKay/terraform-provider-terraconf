package provider

import (
	"context"

	"github.com/RodneyMKay/terraform-provider-terraconf/internal/terraconf"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ function.Function = SanatizeFunction{}
)

func NewSanatizeFunction() function.Function {
	return SanatizeFunction{}
}

type SanatizeFunction struct{}

func (r SanatizeFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "sanatize"
}

func (r SanatizeFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Removes all '_terraconf' tags from an object recursively",
		MarkdownDescription: "When using a data source from this provider to read data, a special '_terraconf' tag is inserted into the data to track metadata about the source of the data. This function can be used to remove those tags, if you want to work with the raw data without any additional metadata. This may be useful if you need to iterate over the keys of an object, for example.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:                "value",
				MarkdownDescription: "Value to sanatize",
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (r SanatizeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var rawValue basetypes.DynamicValue

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &rawValue))
	if resp.Error != nil {
		return
	}

	value, err := terraconf.TerraformDynamicToGoValue(ctx, rawValue)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, err.Error())
	}

	terraconf.RemoveAnnotations(value)

	rawValue, err = terraconf.GoValueToTerraformDynamic(ctx, value)
	if err != nil {
		resp.Error = function.NewFuncError(err.Error())
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, rawValue))
}
