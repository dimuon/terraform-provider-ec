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

package v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Use `self` as value of `observability`'s `deployment_id` attribute
func UseStateForUnknownIfTemplateSame() tfsdk.AttributePlanModifier {
	return useStateForUnknownIfTemplateSame{}
}

type useStateForUnknownIfTemplateSame struct{}

func (r useStateForUnknownIfTemplateSame) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
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

	templateChanged, diags := isAttributeChanged(ctx, path.Root("deployment_template_id"), req)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if templateChanged {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r useStateForUnknownIfTemplateSame) Description(ctx context.Context) string {
	return "Calculate node type value based on current state and `node_roles`'s value."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r useStateForUnknownIfTemplateSame) MarkdownDescription(ctx context.Context) string {
	return "Calculate node type value based on current state and `node_roles`'s value."
}

func isAttributeChanged[V terraformValue](ctx context.Context, p path.Path, req tfsdk.ModifyAttributePlanRequest) (bool, diag.Diagnostics) {
	var planValue V

	if diags := req.Plan.GetAttribute(ctx, p, &planValue); diags.HasError() {
		return false, diags
	}

	var stateValue V

	if diags := req.State.GetAttribute(ctx, p, &stateValue); diags.HasError() {
		return false, diags
	}

	return !planValue.Equal(stateValue), nil
}

type terraformValue interface {
	types.String
	attr.Value
}
