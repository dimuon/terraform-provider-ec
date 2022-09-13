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
	"strings"

	"github.com/blang/semver"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

func modelToState(ctx context.Context, res *models.DeploymentGetResponse, remotes models.RemoteResources, state *Deployment) error {
	var dep Deployment

	if err := dep.fromModel(res, remotes); err != nil {
		return fmt.Errorf("cannot read deployment - %w", err)
	}

	*state = dep

	// if res.Name == nil {
	// 	return fmt.Errorf("server response doesn't contain name")
	// }
	// state.Name.Value = *res.Name

	// state.Alias.Value = res.Alias

	// if res.Metadata != nil {
	// 	state.Tags = flatteners.FlattenTags(res.Metadata.Tags)
	// }

	// if res.Resources != nil {
	// 	var err error

	// 	state.DeploymentTemplateId.Value, err = getDeploymentTemplateID(res.Resources)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	state.Region.Value = getRegion(res.Resources)

	// 	// We're reconciling the version and storing the lowest version of any
	// 	// of the deployment resources. This ensures that if an upgrade fails,
	// 	// the state version will be lower than the desired version, making
	// 	// retries possible. Once more resource types are added, the function
	// 	// needs to be modified to check those as well.
	// 	state.Version.Value, err = getLowestVersion(res.Resources)
	// 	if err != nil {
	// 		// This code path is highly unlikely, but we're bubbling up the
	// 		// error in case one of the versions isn't parseable by semver.
	// 		return fmt.Errorf("failed reading deployment: %w", err)
	// 	}

	// 	esFlattened, err := flattenEsResources(res.Resources.Elasticsearch, *res.Name, remotes)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// state.Elasticsearch =
	// 	if err := d.Set("elasticsearch", esFlattened); err != nil {
	// 		return err
	// 	}

	// 	kibanaFlattened := flattenKibanaResources(res.Resources.Kibana, *res.Name)
	// 	if len(kibanaFlattened) > 0 {
	// 		if err := d.Set("kibana", kibanaFlattened); err != nil {
	// 			return err
	// 		}
	// 	}

	// 	apmFlattened := flattenApmResources(res.Resources.Apm, *res.Name)
	// 	if len(apmFlattened) > 0 {
	// 		if err := d.Set("apm", apmFlattened); err != nil {
	// 			return err
	// 		}
	// 	}

	// 	integrationsServerFlattened := flattenIntegrationsServerResources(res.Resources.IntegrationsServer, *res.Name)
	// 	if len(integrationsServerFlattened) > 0 {
	// 		if err := d.Set("integrations_server", integrationsServerFlattened); err != nil {
	// 			return err
	// 		}
	// 	}

	// 	enterpriseSearchFlattened := flattenEssResources(res.Resources.EnterpriseSearch, *res.Name)
	// 	if len(enterpriseSearchFlattened) > 0 {
	// 		if err := d.Set("enterprise_search", enterpriseSearchFlattened); err != nil {
	// 			return err
	// 		}
	// 	}

	// 	if settings := flattenTrafficFiltering(res.Settings); settings != nil {
	// 		if err := d.Set("traffic_filter", settings); err != nil {
	// 			return err
	// 		}
	// 	}

	// 	if observability := flattenObservability(res.Settings); len(observability) > 0 {
	// 		if err := d.Set("observability", observability); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	return nil
}

func hasRunningResources(res *models.DeploymentGetResponse) bool {
	var hasRunning bool
	if res.Resources != nil {
		for _, r := range res.Resources.Elasticsearch {
			if !isEsResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.Kibana {
			if !isKibanaResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.Apm {
			if !isApmResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.EnterpriseSearch {
			if !isEssResourceStopped(r) {
				hasRunning = true
			}
		}
		for _, r := range res.Resources.IntegrationsServer {
			if !isIntegrationsServerResourceStopped(r) {
				hasRunning = true
			}
		}
	}
	return hasRunning
}

func getDeploymentTemplateID(res *models.DeploymentResources) (string, error) {
	var deploymentTemplateID string
	var foundTemplates []string
	for _, esRes := range res.Elasticsearch {
		if util.IsCurrentEsPlanEmpty(esRes) {
			continue
		}

		var emptyDT = esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate == nil
		if emptyDT {
			continue
		}

		if deploymentTemplateID == "" {
			deploymentTemplateID = *esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID
		}

		foundTemplates = append(foundTemplates,
			*esRes.Info.PlanInfo.Current.Plan.DeploymentTemplate.ID,
		)
	}

	if deploymentTemplateID == "" {
		return "", errors.New("failed to obtain the deployment template id")
	}

	if len(foundTemplates) > 1 {
		return "", fmt.Errorf(
			"there are more than 1 deployment templates specified on the deployment: \"%s\"", strings.Join(foundTemplates, ", "),
		)
	}

	return deploymentTemplateID, nil
}

func getRegion(res *models.DeploymentResources) (region string) {
	for _, r := range res.Elasticsearch {
		if r.Region != nil && *r.Region != "" {
			return *r.Region
		}
	}

	return region
}

func getLowestVersion(res *models.DeploymentResources) (string, error) {
	// We're starting off with a very high version so it can be replaced.
	replaceVersion := `99.99.99`
	version := semver.MustParse(replaceVersion)
	for _, r := range res.Elasticsearch {
		if !util.IsCurrentEsPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.Elasticsearch.Version
			if err := swapLowerVersion(&version, v); err != nil && !isEsResourceStopped(r) {
				return "", fmt.Errorf("elasticsearch version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.Kibana {
		if !util.IsCurrentKibanaPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.Kibana.Version
			if err := swapLowerVersion(&version, v); err != nil && !isKibanaResourceStopped(r) {
				return version.String(), fmt.Errorf("kibana version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.Apm {
		if !util.IsCurrentApmPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.Apm.Version
			if err := swapLowerVersion(&version, v); err != nil && !isApmResourceStopped(r) {
				return version.String(), fmt.Errorf("apm version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.IntegrationsServer {
		if !util.IsCurrentIntegrationsServerPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.IntegrationsServer.Version
			if err := swapLowerVersion(&version, v); err != nil && !isIntegrationsServerResourceStopped(r) {
				return version.String(), fmt.Errorf("integrations_server version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	for _, r := range res.EnterpriseSearch {
		if !util.IsCurrentEssPlanEmpty(r) {
			v := r.Info.PlanInfo.Current.Plan.EnterpriseSearch.Version
			if err := swapLowerVersion(&version, v); err != nil && !isEssResourceStopped(r) {
				return version.String(), fmt.Errorf("enterprise search version '%s' is not semver compliant: %w", v, err)
			}
		}
	}

	if version.String() != replaceVersion {
		return version.String(), nil
	}
	return "", errors.New("Unable to determine the lowest version for any the deployment components")
}

func swapLowerVersion(version *semver.Version, comp string) error {
	if comp == "" {
		return nil
	}

	v, err := semver.Parse(comp)
	if err != nil {
		return err
	}
	if v.LT(*version) {
		*version = v
	}
	return nil
}
