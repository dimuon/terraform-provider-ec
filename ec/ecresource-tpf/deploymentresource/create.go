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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ready(&resp.Diagnostics) {
		return
	}

	var config Deployment
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan Deployment
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := plan.Payload(r.client)
	if err != nil {
		resp.Diagnostics.AddError("cannot create request", err.Error())
		return
	}

	reqID := deploymentapi.RequestID(plan.RequestId.Value)

	res, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:       r.client,
		RequestID: reqID,
		Request:   payload,
		Overrides: &deploymentapi.PayloadOverrides{
			Name:    plan.Name.Value,
			Version: plan.Version.Value,
			Region:  plan.Region.Value,
		},
	})

	if err != nil {
		merr := multierror.NewPrefixed("", err)
		merr.Append(newCreationError(reqID))
		resp.Diagnostics.AddError("failed creating deployment", merr.Error())
		return
	}

	if err := WaitForPlanCompletion(r.client, *res.ID); err != nil {
		merr := multierror.NewPrefixed("", err)
		merr.Append(newCreationError(reqID))
		resp.Diagnostics.AddError("failed tracking create progress", merr.Error())
		return
	}

	tflog.Trace(ctx, "created a resource")

	remoteClustersPayload, diags := ElasticsearchRemoteClusters(plan.Elasticsearch[0].RemoteCluster).Payload(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if err := esremoteclustersapi.Update(esremoteclustersapi.UpdateParams{
		API:             r.client,
		DeploymentID:    *res.ID,
		RefID:           plan.Elasticsearch[0].RefId.Value,
		RemoteResources: remoteClustersPayload,
	}); err != nil {
		resp.Diagnostics.AddError("failed updating remote cluster", err.Error())
	}

	deployment, err := r.read(ctx, *res.ID, plan)

	if err != nil {
		resp.Diagnostics.AddError("Read error", err.Error())
	}

	if deployment == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if err := deployment.ParseCredentials(res.Resources); err != nil {
		resp.Diagnostics.AddError("failed parse credentials", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deployment)...)
}

func newCreationError(reqID string) error {
	return fmt.Errorf(
		`set "request_id" to "%s" to recreate the deployment resources`, reqID,
	)
}
