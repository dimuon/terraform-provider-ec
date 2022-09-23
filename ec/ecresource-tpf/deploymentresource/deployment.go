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

	"github.com/blang/semver"
	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi/deptemplateapi"
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/elastic/terraform-provider-ec/ec/internal/converters"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Deployment struct {
	Id                    types.String        `tfsdk:"id"`
	Alias                 types.String        `tfsdk:"alias"`
	Version               types.String        `tfsdk:"version"`
	Region                types.String        `tfsdk:"region"`
	DeploymentTemplateId  types.String        `tfsdk:"deployment_template_id"`
	Name                  types.String        `tfsdk:"name"`
	RequestId             types.String        `tfsdk:"request_id"`
	ElasticsearchUsername types.String        `tfsdk:"elasticsearch_username"`
	ElasticsearchPassword types.String        `tfsdk:"elasticsearch_password"`
	ApmSecretToken        types.String        `tfsdk:"apm_secret_token"`
	TrafficFilter         []string            `tfsdk:"traffic_filter"`
	Tags                  map[string]string   `tfsdk:"tags"`
	Elasticsearch         Elasticsearches     `tfsdk:"elasticsearch"`
	Kibana                Kibanas             `tfsdk:"kibana"`
	Apm                   Apms                `tfsdk:"apm"`
	IntegrationsServer    IntegrationsServers `tfsdk:"integrations_server"`
	EnterpriseSearch      EnterpriseSearches  `tfsdk:"enterprise_search"`
	Observability         Observabilities     `tfsdk:"observability"`
}

var (
	dataTiersVersion = semver.MustParse("7.10.0")
)

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

	dep.Tags = converters.TagsToMap(res.Metadata.Tags)

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

func (d *Deployment) Payload(client *api.API) (*models.DeploymentCreateRequest, error) {
	var result = models.DeploymentCreateRequest{
		Name:      d.Name.Value,
		Alias:     d.Alias.Value,
		Resources: &models.DeploymentCreateResources{},
		Settings:  &models.DeploymentCreateSettings{},
		Metadata:  &models.DeploymentCreateMetadata{},
	}

	dtID := d.DeploymentTemplateId.Value
	version := d.Version.Value

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     d.Region.Value,
		HideInstanceConfigurations: true,
	})
	if err != nil {
		return nil, err
	}

	useNodeRoles, err := compatibleWithNodeRoles(version)
	if err != nil {
		return nil, err
	}

	merr := multierror.NewPrefixed("invalid configuration")

	esRes, err := d.Elasticsearch.Payload(template, dtID, version, useNodeRoles)
	// enrichElasticsearchTemplate(
	// 	esResource(template), dtID, version, useNodeRoles,
	// ),

	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := d.Kibana.Payload(template)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apmRes, err := d.Apm.Payload(template)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Apm = append(result.Resources.Apm, apmRes...)

	integrationsServerRes, err := d.IntegrationsServer.Payload(template)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerRes...)

	enterpriseSearchRes, err := d.EnterpriseSearch.Payload(template)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	if err := merr.ErrorOrNil(); err != nil {
		return nil, err
	}

	trafficFilterToModel(d.TrafficFilter, &result)

	observability, err := d.Observability.Model(client)
	if err != nil {
		return nil, err
	}
	result.Settings.Observability = observability

	result.Metadata.Tags = converters.MapToTags(d.Tags)

	return &result, nil
}

// parseCredentials parses the Create or Update response Resources populating
// credential settings in the Terraform state if the keys are found, currently
// populates the following credentials in plain text:
// * Elasticsearch username and Password
func (dep *Deployment) ParseCredentials(resources []*models.DeploymentResource) error {

	for _, res := range resources {

		if creds := res.Credentials; creds != nil {
			if creds.Username != nil && *creds.Username != "" {
				dep.ElasticsearchUsername.Value = *creds.Username
			}

			if creds.Password != nil && *creds.Password != "" {
				dep.ElasticsearchPassword.Value = *creds.Password
			}
		}

		if res.SecretToken != "" {
			dep.ApmSecretToken.Value = res.SecretToken
		}
	}

	return nil
}

func NewTrafficFilters(in *models.DeploymentSettings) ([]string, error) {
	if in == nil || in.TrafficFilterSettings == nil || len(in.TrafficFilterSettings.Rulesets) == 0 {
		return nil, nil
	}

	var rules []string

	return append(rules, in.TrafficFilterSettings.Rulesets...), nil
}

func compatibleWithNodeRoles(version string) (bool, error) {
	deploymentVersion, err := semver.Parse(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse Elasticsearch version: %w", err)
	}

	return deploymentVersion.GE(dataTiersVersion), nil
}

// trafficFilterToModel expands the flattened "traffic_filter" settings to
// a DeploymentCreateRequest.
func trafficFilterToModel(set []string, req *models.DeploymentCreateRequest) {
	if set == nil || req == nil {
		return
	}

	if len(set) == 0 {
		return
	}

	if req.Settings == nil {
		req.Settings = &models.DeploymentCreateSettings{}
	}

	if req.Settings.TrafficFilterSettings == nil {
		req.Settings.TrafficFilterSettings = &models.TrafficFilterSettings{}
	}

	req.Settings.TrafficFilterSettings.Rulesets = append(
		req.Settings.TrafficFilterSettings.Rulesets,
		set...,
	)
}
