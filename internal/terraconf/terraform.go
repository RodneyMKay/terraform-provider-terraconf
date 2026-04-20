package terraconf

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// In the provider internal code, we use a representation that is closer to
// json/yaml style values (map[string]any, []any, boolean, int64, float64
// and string), which is more convenient to work with in Go. However,
// to pass them to terraform, we need to convert them to/from terraform values.
// The functions in this file facilitate this conversion.

// Converts Go objects to a terraform DynamicValue.
func GoValueToTerraformDynamic(ctx context.Context, val any) (basetypes.DynamicValue, error) {
	attrVal, err := GoValueToTerraformAttr(ctx, val)
	if err != nil {
		return types.DynamicNull(), err
	}
	return types.DynamicValue(attrVal), nil
}

// Converts a terraform DynamicValue to Go objects.
func TerraformDynamicToGoValue(ctx context.Context, val basetypes.DynamicValue) (any, error) {
	if val.IsNull() || val.IsUnknown() {
		return nil, nil
	}
	return TerraformAttrToGoValue(ctx, val.UnderlyingValue())
}

// Converts Go objects to a terraform attr.Value.
func GoValueToTerraformAttr(ctx context.Context, val any) (attr.Value, error) {
	if val == nil {
		return types.DynamicNull(), nil
	}

	switch v := val.(type) {
	case string:
		return types.StringValue(v), nil
	case bool:
		return types.BoolValue(v), nil
	case int:
		return types.Int64Value(int64(v)), nil
	case int64:
		return types.Int64Value(v), nil
	case float64:
		return types.Float64Value(v), nil
	case []any:
		// JSON arrays can have mixed types, so we use a Tuple instead of a List
		elements := make([]attr.Value, len(v))
		elemTypes := make([]attr.Type, len(v))
		for i, item := range v {
			attrVal, err := GoValueToTerraformAttr(ctx, item)
			if err != nil {
				return nil, err
			}
			elements[i] = attrVal
			elemTypes[i] = attrVal.Type(ctx)
		}

		tuple, diags := types.TupleValue(elemTypes, elements)
		if diags.HasError() {
			return nil, fmt.Errorf("error creating tuple: %v", diags)
		}
		return tuple, nil
	case map[string]any:
		// JSON objects can have mixed types, so we use an Object instead of
		// a Map
		attrTypes := make(map[string]attr.Type)
		attrValues := make(map[string]attr.Value)
		for key, item := range v {
			attrVal, err := GoValueToTerraformAttr(ctx, item)
			if err != nil {
				return nil, err
			}
			attrTypes[key] = attrVal.Type(ctx)
			attrValues[key] = attrVal
		}

		obj, diags := types.ObjectValue(attrTypes, attrValues)
		if diags.HasError() {
			return nil, fmt.Errorf("error creating object: %v", diags)
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("unsupported go type for conversion: %T", v)
	}
}

// Converts a terraform attr.Value to Go objects.
func TerraformAttrToGoValue(ctx context.Context, val attr.Value) (any, error) {
	if val.IsNull() || val.IsUnknown() {
		return nil, nil
	}

	switch v := val.(type) {
	case basetypes.StringValue:
		return v.ValueString(), nil
	case basetypes.BoolValue:
		return v.ValueBool(), nil
	case basetypes.Int64Value:
		return v.ValueInt64(), nil
	case basetypes.Float64Value:
		return v.ValueFloat64(), nil
	case basetypes.NumberValue:
		// In terraform core this case can only be a big.Float from math/big.
		// This extra precision is not implemented by any of our storage
		// backends, so we convert it to a regular float64.
		f, _ := v.ValueBigFloat().Float64()
		return f, nil
	case basetypes.ListValue:
		var result []any
		for _, elem := range v.Elements() {
			parsed, err := TerraformAttrToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, parsed)
		}
		return result, nil
	case basetypes.TupleValue:
		var result []any
		for _, elem := range v.Elements() {
			parsed, err := TerraformAttrToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result = append(result, parsed)
		}
		return result, nil
	case basetypes.MapValue:
		result := make(map[string]any)
		for key, elem := range v.Elements() {
			parsed, err := TerraformAttrToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result[key] = parsed
		}
		return result, nil
	case basetypes.ObjectValue:
		result := make(map[string]any)
		for key, elem := range v.Attributes() {
			parsed, err := TerraformAttrToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result[key] = parsed
		}
		return result, nil
	case basetypes.DynamicValue:
		// Unwrap the dynamic value and process the underlying value recursively
		return TerraformAttrToGoValue(ctx, v.UnderlyingValue())
	default:
		return nil, fmt.Errorf("unsupported attr.Value type for conversion: %T", v)
	}
}
