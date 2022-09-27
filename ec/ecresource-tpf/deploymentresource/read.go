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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.ready(&resp.Diagnostics) {
		return
	}

	var curState Deployment

	diags := req.State.Get(ctx, &curState)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var newState *Deployment
	var err error

	if newState, err = r.read(ctx, curState.Id.Value, curState); err != nil {
		resp.Diagnostics.AddError("Read error", err.Error())
	}

	if newState == nil {
		resp.State.RemoveResource(ctx)
	}

	if newState != nil {
		diags = resp.State.Set(ctx, newState)
	}

	resp.Diagnostics.Append(diags...)
}

func (r Resource) read(ctx context.Context, id string, state Deployment) (*Deployment, error) {
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
			return nil, err
		}
		return nil, fmt.Errorf("failed reading deployment - %w", err)
	}

	if !hasRunningResources(res) {
		return nil, nil
	}

	remotes, err := esremoteclustersapi.Get(esremoteclustersapi.GetParams{
		API: r.client, DeploymentID: id,
		RefID: state.Elasticsearch[0].RefId.Value,
	})
	if err != nil {
		return nil, fmt.Errorf("failed reading remote clusters - %w", err)
	}
	if remotes == nil {
		remotes = &models.RemoteResources{}
	}

	if dep, err := NewDeployment(res, remotes); err != nil {
		return nil, err
	} else {
		return dep, nil
	}
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
