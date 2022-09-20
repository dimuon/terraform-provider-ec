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
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/terraform-provider-ec/ec/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Deployment struct {
	Id                    types.String          `tfsdk:"id"`
	Alias                 types.String          `tfsdk:"alias"`
	Version               types.String          `tfsdk:"version"`
	Region                types.String          `tfsdk:"region"`
	DeploymentTemplateId  types.String          `tfsdk:"deployment_template_id"`
	Name                  types.String          `tfsdk:"name"`
	RequestId             types.String          `tfsdk:"request_id"`
	ElasticsearchUsername types.String          `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword types.String          `tfsdk:"elasticsearch_password"`
	ApmSecretToken        types.String          `tfsdk:"apm_secret_token"`
	TrafficFilter         []string              `tfsdk:"traffic_filter"`
	Tags                  map[string]string     `tfsdk:"tags"`
	Elasticsearch         []*Elasticsearch      `tfsdk:"elasticsearch"`
	Kibana                []*Kibana             `tfsdk:"kibana"`
	Apm                   []*Apm                `tfsdk:"apm"`
	IntegrationsServer    []*IntegrationsServer `tfsdk:"integrations_server"`
	EnterpriseSearch      []*EnterpriseSearch   `tfsdk:"enterprise_search"`
	Observability         []*Observability      `tfsdk:"observability"`
}

func missingField(field string) error {
	return fmt.Errorf("server response doesn't contain deployment '%s'", field)
}

func NewDeployment(res *models.DeploymentGetResponse, remotes *models.RemoteResources) (*Deployment, error) {
	var dep Deployment

	if res.ID == nil {
		return nil, missingField("ID")
	}
	dep.Id.Value = *res.ID

	dep.Alias.Value = res.Alias

	if res.Name == nil {
		return nil, missingField("Name")
	}
	dep.Name.Value = *res.Name

	dep.Tags = flatteners.TagsToMap(res.Metadata.Tags)

	if res.Resources == nil {
		return nil, nil
	}

	var err error

	dep.DeploymentTemplateId.Value, err = getDeploymentTemplateID(res.Resources)
	if err != nil {
		return nil, err
	}

	dep.Region.Value = getRegion(res.Resources)

	// We're reconciling the version and storing the lowest version of any
	// of the deployment resources. This ensures that if an upgrade fails,
	// the state version will be lower than the desired version, making
	// retries possible. Once more resource types are added, the function
	// needs to be modified to check those as well.
	dep.Version.Value, err = getLowestVersion(res.Resources)
	if err != nil {
		// This code path is highly unlikely, but we're bubbling up the
		// error in case one of the versions isn't parseable by semver.
		return nil, fmt.Errorf("failed reading deployment: %w", err)
	}

	if dep.Elasticsearch, err = NewElasticsearches(res.Resources.Elasticsearch, remotes); err != nil {
		return nil, err
	}

	if dep.Kibana, err = NewKibanas(res.Resources.Kibana); err != nil {
		return nil, err
	}

	if dep.Apm, err = NewApms(res.Resources.Apm); err != nil {
		return nil, err
	}

	if dep.IntegrationsServer, err = NewIntegrationsServers(res.Resources.IntegrationsServer); err != nil {
		return nil, err
	}

	if dep.EnterpriseSearch, err = NewEnterpriseSearches(res.Resources.EnterpriseSearch); err != nil {
		return nil, err
	}

	if dep.EnterpriseSearch, err = NewEnterpriseSearches(res.Resources.EnterpriseSearch); err != nil {
		return nil, err
	}

	if dep.TrafficFilter, err = NewTrafficFilters(res.Settings); err != nil {
		return nil, err
	}

	if dep.Observability, err = NewObservability(res.Settings); err != nil {
		return nil, err
	}
	return &dep, nil
}

type ElasticsearchSnapshotSource struct {
	SourceElasticsearchClusterId types.String `tfsdk:"source_elasticsearch_cluster_id"`
	SnapshotName                 types.String `tfsdk:"snapshot_name"`
}

type ElasticsearchStrategy struct {
	Type types.String `tfsdk:"type"`
}

func NewTrafficFilters(in *models.DeploymentSettings) ([]string, error) {
	if in == nil || in.TrafficFilterSettings == nil || len(in.TrafficFilterSettings.Rulesets) == 0 {
		return nil, nil
	}

	var rules []string

	return append(rules, in.TrafficFilterSettings.Rulesets...), nil
}
