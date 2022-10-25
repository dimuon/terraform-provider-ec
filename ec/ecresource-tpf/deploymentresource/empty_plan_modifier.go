// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deploymentresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Use `self` as value of `observability`'s `deployment_id` attribute
func UseStateIfEmptyCollection() tfsdk.AttributePlanModifier {
	return useStateIfEmptyCollection{}
}

type useStateIfEmptyCollection struct{}

func (r useStateIfEmptyCollection) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	// // if we have no state value, there's nothing to preserve
	// if req.AttributeState.IsNull() {
	// 	return
	// }

	// if !resp.AttributePlan.IsUnknown() {
	// 	return
	// }

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.AttributeConfig.IsUnknown() {
		return
	}

	if _, ok := req.AttributePlan.Type(ctx).(attr.TypeWithElementType); !ok {
		return
	}

	val, _ := req.AttributePlan.ToTerraformValue(ctx)

	switch val.Type().(type) {
	case tftypes.List, tftypes.Set:
		var vals []tftypes.Value
		if err := val.As(&vals); err != nil {
			return
		}
		if len(vals) == 0 {
			resp.AttributePlan = req.AttributeState
			return
		}
		return
	default:
		return
	}
}

// Description returns a human-readable description of the plan modifier.
func (r useStateIfEmptyCollection) Description(ctx context.Context) string {
	return "Use state if it's empty collection."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r useStateIfEmptyCollection) MarkdownDescription(ctx context.Context) string {
	return "Use state if it's empty collection"
}
