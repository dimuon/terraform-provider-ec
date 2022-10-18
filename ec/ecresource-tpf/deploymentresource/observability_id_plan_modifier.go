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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Use `self` as value of `observability`'s `deployment_id` attribute
func UseSelfForObservabilityId() tfsdk.AttributePlanModifier {
	return useSelfForObservabilityId{}
}

type useSelfForObservabilityId struct{}

func (r useSelfForObservabilityId) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
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

	if !req.AttributeState.Equal(types.String{Value: "self"}) {
		return
	}

	var deploymentId types.String

	diags := req.Config.GetAttribute(ctx, path.Root("id"), &deploymentId)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if deploymentId.IsUnknown() || deploymentId.IsNull() {
		return
	}

	var observabilityDeploymentId types.String

	diags = tfsdk.ValueAs(ctx, req.AttributePlan, &observabilityDeploymentId)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if observabilityDeploymentId.IsUnknown() || observabilityDeploymentId.IsNull() {
		return
	}

	if observabilityDeploymentId == deploymentId {
		resp.AttributePlan = req.AttributeState
	}
}

// Description returns a human-readable description of the plan modifier.
func (r useSelfForObservabilityId) Description(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r useSelfForObservabilityId) MarkdownDescription(ctx context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}
