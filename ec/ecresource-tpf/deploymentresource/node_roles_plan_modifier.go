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
func UseNodeRolesDefault() tfsdk.AttributePlanModifier {
	return nodeRolesDefault{}
}

type nodeRolesDefault struct{}

func (r nodeRolesDefault) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
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

	deploymentVersionChanged, diags := isAttributeChanged(ctx, path.Root("version"), req)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if deploymentVersionChanged {
		return
	}

	topologyElementChanged, diags := isElasticTopologyChanged(ctx, req.AttributePath.ParentPath().ParentPath(), req)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if topologyElementChanged {
		return
	}

	if !req.AttributeState.IsUnknown() && !req.AttributeState.IsNull() {
		resp.AttributePlan = req.AttributeState
		return
	}

	nodeRolesPath := req.AttributePath.ParentPath().AtName("node_type_data")

	var nodeTypeState types.String

	diags = req.State.GetAttribute(ctx, nodeRolesPath, &nodeTypeState)

	resp.Diagnostics = append(resp.Diagnostics, diags...)

	if diags.HasError() {
		return
	}

	if nodeTypeState.IsUnknown() || nodeTypeState.IsNull() {
		return
	}

	resp.AttributePlan = req.AttributeState
}

// Description returns a human-readable description of the plan modifier.
func (r nodeRolesDefault) Description(ctx context.Context) string {
	return "Calculate node roles value based on current state and `node_type_data`'s value."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r nodeRolesDefault) MarkdownDescription(ctx context.Context) string {
	return "Calculate node roles value based on current state and `node_type_data`'s value."
}
