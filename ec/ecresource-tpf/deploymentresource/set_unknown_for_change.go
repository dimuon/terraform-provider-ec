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

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UseUnknownTopologyBlockIfEmpty() tfsdk.AttributePlanModifier {
	return useUnknownTopologyBlockIfEmpty{}
}

type useUnknownTopologyBlockIfEmpty struct{}

// Modify copies the attribute's prior state to the attribute plan if the prior
// state value is not null.
func (r useUnknownTopologyBlockIfEmpty) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if resp.AttributePlan.IsUnknown() {
		return
	}

	var esList types.List

	if diags := tfsdk.ValueAs(ctx, resp.AttributePlan, &esList); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if esList.IsNull() || len(esList.Elems) == 0 {
		resp.AttributePlan = types.List{
			Unknown:  true,
			ElemType: elasticsearchTopologySchema().Type().(types.ListType).ElemType,
		}
	}
}

// Description returns a human-readable description of the plan modifier.
func (r useUnknownTopologyBlockIfEmpty) Description(ctx context.Context) string {
	return ""
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r useUnknownTopologyBlockIfEmpty) MarkdownDescription(ctx context.Context) string {
	return ""
}
