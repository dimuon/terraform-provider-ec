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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Use `self` as value of `observability`'s `deployment_id` attribute
func UseApmTokenDefault() tfsdk.AttributePlanModifier {
	return apmTokenDefault{}
}

type apmTokenDefault struct{}

func (r apmTokenDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeState == nil || resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if !resp.AttributePlan.IsUnknown() {
		return
	}

	// if the config is the unknown value, use the unknown value otherwise, interpolation gets messed up
	if req.AttributeConfig.IsUnknown() {
		return
	}

	apmChanged, diags := isApmChanged(ctx, path.Root("apm"), req)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if apmChanged {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r apmTokenDefault) Description(ctx context.Context) string {
	return "Calculate apm token value based on changes in configuration."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r apmTokenDefault) MarkdownDescription(ctx context.Context) string {
	return "Calculate apm token value based on changes in configuration."
}

func isApmChanged(ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
	var planValue types.List

	if diags := req.Plan.GetAttribute(ctx, p, &planValue); diags.HasError() {
		return false, diags
	}

	if planValue.IsUnknown() {
		return false, nil
	}

	var plan []ApmTF

	if diags := planValue.ElementsAs(ctx, &plan, true); diags.HasError() {
		return false, diags
	}

	var stateValue types.List

	if diags := req.State.GetAttribute(ctx, p, &stateValue); diags.HasError() {
		return false, diags
	}

	var state []ApmTF

	if diags := stateValue.ElementsAs(ctx, &state, true); diags.HasError() {
		return false, diags
	}

	if len(plan) == 0 {
		return false, nil
	}

	if len(state) == 0 {
		return true, nil
	}

	return true, nil
}
