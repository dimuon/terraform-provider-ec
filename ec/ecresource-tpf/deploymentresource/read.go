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
	"errors"

	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.ready(&resp.Diagnostics) {
		return
	}

	var curState DeploymentTF

	diags := req.State.Get(ctx, &curState)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var newState *DeploymentTF
	var err error

	if newState, diags = r.read(ctx, curState.Id.Value, curState); err != nil {
		resp.Diagnostics.Append(diags...)
	}

	if newState == nil {
		resp.State.RemoveResource(ctx)
	}

	if newState != nil {
		diags = resp.State.Set(ctx, newState)
	}

	resp.Diagnostics.Append(diags...)
}

func (r Resource) read(ctx context.Context, id string, state DeploymentTF) (*DeploymentTF, diag.Diagnostics) {
	var diags diag.Diagnostics
	res, err := deploymentapi.Get(deploymentapi.GetParams{
		API: r.client, DeploymentID: id,
		QueryParams: deputil.QueryParams{
			ShowSettings:     true,
			ShowPlans:        true,
			ShowMetadata:     true,
			ShowPlanDefaults: true,
		},
	})
	if err != nil {
		if deploymentNotFound(err) {
			diags.AddError("Deployment not found", err.Error())
			return nil, diags
		}
		diags.AddError("Deloyment get error", err.Error())
		return nil, diags
	}

	if !hasRunningResources(res) {
		return nil, nil
	}

	var es ElasticsearchTF
	if diags := tfsdk.ValueAs(ctx, state.Elasticsearch.Elems[0], &es); diags.HasError() {
		return nil, diags
	}

	remotes, err := esremoteclustersapi.Get(esremoteclustersapi.GetParams{
		API: r.client, DeploymentID: id,
		RefID: es.RefId.Value,
	})
	if err != nil {
		diags.AddError("Remote clusters read error", err.Error())
		return nil, diags
	}
	if remotes == nil {
		remotes = &models.RemoteResources{}
	}

	dep, err := readDeployment(res, remotes)
	if err != nil {
		diags.AddError("Deployment read error", err.Error())
		return nil, diags
	}

	var deployment DeploymentTF
	schema, diags := r.GetSchema(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if diags := tfsdk.ValueFrom(ctx, dep, schema.Type(), &deployment); diags.HasError() {
		return nil, diags
	}

	return &deployment, diags
}

func deploymentNotFound(err error) bool {
	// We're using the As() call since we do not care about the error value
	// but do care about the error's contents type since it's an implicit 404.
	var notDeploymentNotFound *deployments.GetDeploymentNotFound
	if errors.As(err, &notDeploymentNotFound) {
		return true
	}

	// We also check for the case where a 403 is thrown for ESS.
	return apierror.IsRuntimeStatusCode(err, 403)
}
