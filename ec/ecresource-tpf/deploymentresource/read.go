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

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/apierror"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deputil"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/esremoteclustersapi"
	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func read(ctx context.Context, client *api.API, state *Deployment) error {
	res, err := deploymentapi.Get(deploymentapi.GetParams{
		API: client, DeploymentID: state.Id.Value,
		QueryParams: deputil.QueryParams{
			ShowSettings:     true,
			ShowPlans:        true,
			ShowMetadata:     true,
			ShowPlanDefaults: true,
		},
	})
	if err != nil {
		if deploymentNotFound(err) {
			state.Id.Value = ""
			return nil
		}
		return fmt.Errorf("failed reading deployment - %w", err)
	}

	if !hasRunningResources(res) {
		state.Id.Value = ""
		return nil
	}

	remotes, err := esremoteclustersapi.Get(esremoteclustersapi.GetParams{
		API: client, DeploymentID: state.Id.Value,
		RefID: state.Elasticsearch[0].RefId.Value,
	})
	if err != nil {
		return fmt.Errorf("failed reading remote clusters - %w", err)
	}
	if remotes == nil {
		remotes = &models.RemoteResources{}
	}

	if err := modelToState(ctx, res, remotes, state); err != nil {
		return err
	}

	return nil
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
