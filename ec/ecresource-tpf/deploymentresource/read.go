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

	"github.com/blang/semver"
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

func (r Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if !r.ready(&response.Diagnostics) {
		return
	}

	var curState DeploymentTF

	diags := request.State.Get(ctx, &curState)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var newState *DeploymentTF
	var err error

	if newState, diags = r.read(ctx, curState.Id.Value, curState, nil); err != nil {
		response.Diagnostics.Append(diags...)
	}

	if newState == nil {
		response.State.RemoveResource(ctx)
	}

	if newState != nil {
		diags = response.State.Set(ctx, newState)
	}

	response.Diagnostics.Append(diags...)
}

func (r Resource) read(ctx context.Context, id string, current DeploymentTF, deploymentResources []*models.DeploymentResource) (*DeploymentTF, diag.Diagnostics) {
	var diags diag.Diagnostics

	response, err := deploymentapi.Get(deploymentapi.GetParams{
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

	if response.Resources == nil || len(response.Resources.Elasticsearch) == 0 {
		diags.AddError("Get resource error", "cannot find Elasticsearch in response resources")
		return nil, diags
	}

	if response.Resources.Elasticsearch[0].Info.PlanInfo.Current != nil && response.Resources.Elasticsearch[0].Info.PlanInfo.Current.Plan != nil {
		if err := checkVersion(response.Resources.Elasticsearch[0].Info.PlanInfo.Current.Plan.Elasticsearch.Version); err != nil {
			diags.AddError("Get resource error", err.Error())
			return nil, diags
		}
	}

	if !hasRunningResources(response) {
		return nil, nil
	}

	refId := ""

	var elasticsearch *ElasticsearchTF

	if diags = getFirst(ctx, current.Elasticsearch, &elasticsearch); diags.HasError() {
		return nil, diags
	}

	if elasticsearch != nil {
		refId = elasticsearch.RefId.Value
	}

	remotes, err := esremoteclustersapi.Get(esremoteclustersapi.GetParams{
		API: r.client, DeploymentID: id,
		RefID: refId,
	})
	if err != nil {
		diags.AddError("Remote clusters read error", err.Error())
		return nil, diags
	}
	if remotes == nil {
		remotes = &models.RemoteResources{}
	}

	deployment, err := readDeployment(response, remotes, deploymentResources)
	if err != nil {
		diags.AddError("Deployment read error", err.Error())
		return nil, diags
	}

	deployment.RequestId = current.RequestId.Value

	if current.ElasticsearchPassword.Value != "" {
		deployment.ElasticsearchPassword = current.ElasticsearchPassword.Value
	}

	if current.ElasticsearchUsername.Value != "" {
		deployment.ElasticsearchUsername = current.ElasticsearchUsername.Value
	}

	if current.ApmSecretToken.Value != "" {
		deployment.ApmSecretToken = current.ApmSecretToken.Value
	}

	var deploymentTF DeploymentTF
	schema, diags := r.GetSchema(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if diags := tfsdk.ValueFrom(ctx, deployment, schema.Type(), &deploymentTF); diags.HasError() {
		return nil, diags
	}

	return &deploymentTF, diags
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

// Setting this variable here so that it is parsed at compile time in case
// any errors are thrown, they are at compile time not when the user runs it.
var minimumSupportedVersion = semver.MustParse("6.6.0")

func checkVersion(version string) error {
	v, err := semver.New(version)

	if err != nil {
		return fmt.Errorf("unable to parse deployment version: %w", err)
	}

	if v.LT(minimumSupportedVersion) {
		return fmt.Errorf(
			`invalid deployment version "%s": minimum supported version is "%s"`,
			v.String(), minimumSupportedVersion.String(),
		)
	}

	return nil
}
