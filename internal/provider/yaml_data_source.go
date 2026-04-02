package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/RodneyMKay/terraform-provider-terraconf/internal/terraconf"
)

var _ datasource.DataSource = &YamlDataSource{}

func NewYamlDataSource() datasource.DataSource {
	return &YamlDataSource{}
}

type YamlDataSource struct {
}

type YamlDataSourceModel struct {
	Glob           types.String  `tfsdk:"input_glob"`
	Schema         types.String  `tfsdk:"schema_file"`
	AutoFlatten    types.Bool    `tfsdk:"auto_flatten"`
	AddAnnotations types.Bool    `tfsdk:"add_annotations"`
	Output         types.Dynamic `tfsdk:"output"`
}

func (d *YamlDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_yaml"
}

func (d *YamlDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "YAML file data source. Allows reading configuration based on one or more YAML files. The filepaths are specified as a GLOB pattern. Optionally, a JSON schema can be specified to validate the configuration against.",

		Attributes: map[string]schema.Attribute{
			"input_glob": schema.StringAttribute{
				MarkdownDescription: "GLOB pattern to match YAML files against. For example, `./config/*.yaml`.",
				Required:            true,
			},
			"schema_file": schema.StringAttribute{
				MarkdownDescription: "Path to a JSON schema file to validate the YAML files against. If validation fails, returns an error with the location in the YAML file.",
				Optional:            true,
			},
			"auto_flatten": schema.BoolAttribute{
				MarkdownDescription: "If true (default), automatically flattens lists at the root of each YAML file into the output list. If false, each YAML file's content is added as a single element.",
				Optional:            true,
			},
			"add_annotations": schema.BoolAttribute{
				MarkdownDescription: "If true (default), a special _terraconf annotation is added to every object in the YAML structure, whcih encodes the location of that object inside the configuration. This can be used in a custom function to display a rich error message, which points to the exact location of the error in the YAML file.",
				Optional:            true,
			},
			"output": schema.DynamicAttribute{
				MarkdownDescription: "The output of the data source. The structure of this output depends on the YAML files read. Returns a list where each element corresponds to one YAML file (or flattened list items if auto_flatten is true).",
				Computed:            true,
			},
		},
	}
}

func (d *YamlDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *YamlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read Terraform configuration data into the model
	var data YamlDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the required input parameters
	glob := data.Glob.ValueString()
	autoFlatten := data.AutoFlatten.IsNull() || data.AutoFlatten.ValueBool()
	addAnnotations := data.AddAnnotations.IsNull() || data.AddAnnotations.ValueBool()
	schemaPath := data.Schema.ValueString()

	// Read all of the files
	files, err := terraconf.FindGlobFiles(glob)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Finding YAML Files",
			"An error was encountered when trying to find YAML files matching the provided glob pattern. "+
				"Please verify the glob pattern is correct and that the files exist.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	yamlValues := make([]any, 0, len(files))

	for _, file := range files {
		tflog.Trace(ctx, fmt.Sprintf("Reading YAML file: %s", file))
		
		// Load YAML without annotations first for schema validation
		value, err := terraconf.LoadYAML(file)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Loading YAML File",
				fmt.Sprintf("An error was encountered when trying to load YAML file %s: %s. This is often due to a complex key being used in the YAML structure (like an object being used as the key for an entry in that object).", file, err.Error()),
			)
			return
		}

		// Validate against JSON schema if provided
		if schemaPath != "" {
			schemaErrors, err := terraconf.CheckWithSchema(value, schemaPath)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Reading JSON Schema",
					fmt.Sprintf("An error was encountered when trying to read schema %s: %s", schemaPath, err.Error()),
				)
				return
			}

			for _, schemaError := range schemaErrors {
				detailMessage, err := terraconf.TraceError(file, schemaError.Pointer, schemaError.Message)
				if err != nil {
					resp.Diagnostics.AddError(
						"Failed to Trace YAML Schema Validation Error",
						fmt.Sprintf("An error was encountered when trying to trace a schema validation error for file %s at location %s: %s", file, schemaError.Pointer.Raw(), err.Error()),
					)
					return
				}

				resp.Diagnostics.AddError(
					"YAML Schema Validation Error",
					detailMessage,
				)
			}
		}

		// Add annotations
		if addAnnotations {
			err = terraconf.AddAnnotations(value, file)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Annotating YAML Data",
					fmt.Sprintf("An error was encountered when trying to annotate the data from file %s: %s", file, err.Error()),
				)
				return
			}
		}

		// Automatically flatten lists
		// FIXME: We should make this more intelligent by only flattening
		//  if all elements of the top-level list are lists themselves
		if autoFlatten {
			if valueList, ok := value.([]any); ok {
				yamlValues = append(yamlValues, valueList...)
				continue
			}
		}

		yamlValues = append(yamlValues, value)
	}

	// Write the output
	terraformValue, err := terraconf.GoValueToTerraformDynamic(ctx, yamlValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting YAML to Terraform Value",
			fmt.Sprintf("An error was encountered when trying to convert the loaded YAML data into a format that can be used in Terraform: %s", err.Error()),
		)
		return
	}

	data.Output = terraformValue

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
