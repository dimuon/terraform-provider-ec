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

package planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Use `self` as value of `observability`'s `deployment_id` attribute
func UseStateForNoChange() tfsdk.AttributePlanModifier {
	return useStateForNoChange{}
}

type useStateForNoChange struct{}

func (r useStateForNoChange) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	// if we have no state value, there's nothing to preserve
	// if req.AttributeState.IsNull() {
	// 	return
	// }

	// if it's not planned to be the unknown value, stick with the concrete plan
	if !resp.AttributePlan.IsUnknown() {
		return
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.AttributeConfig.IsUnknown() {
		return
	}

	diffs, err := req.Plan.Raw.Diff(req.State.Raw)

	if err != nil {
		resp.Diagnostics.AddError("cannot get diff between plan and state values", err.Error())
		return
	}

	for _, dif := range diffs {

		if dif.Value1 == nil {
			continue
		}

		if dif.Value1.IsKnown() {
			return
		}
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r useStateForNoChange) Description(ctx context.Context) string {
	return "Use state value if there is no change in config."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r useStateForNoChange) MarkdownDescription(ctx context.Context) string {
	return "Use state value if there is no change in config."
}
