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
	Apm                   types.List          `tfsdk:"apm"`
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
	ctx := context.TODO()
	var dep Deployment

	if res.ID == nil {
		return nil, missingField("ID")
	}
	dep.Id = types.String{Value: *res.ID}

	dep.Alias = types.String{Value: res.Alias}

	if res.Name == nil {
		return nil, missingField("Name")
	}
	dep.Name = types.String{Value: *res.Name}

	dep.Tags = converters.TagsToMap(res.Metadata.Tags)

	if res.Resources == nil {
		return nil, nil
	}

	templateID, err := getDeploymentTemplateID(res.Resources)
	if err != nil {
		return nil, err
	}

	dep.DeploymentTemplateId = types.String{Value: templateID}

	dep.Region = types.String{Value: getRegion(res.Resources)}

	// We're reconciling the version and storing the lowest version of any
	// of the deployment resources. This ensures that if an upgrade fails,
	// the state version will be lower than the desired version, making
	// retries possible. Once more resource types are added, the function
	// needs to be modified to check those as well.
	version, err := getLowestVersion(res.Resources)
	if err != nil {
		// This code path is highly unlikely, but we're bubbling up the
		// error in case one of the versions isn't parseable by semver.
		return nil, fmt.Errorf("failed reading deployment: %w", err)
	}
	dep.Version = types.String{Value: version}

	if dep.Elasticsearch, err = NewElasticsearches(res.Resources.Elasticsearch, remotes); err != nil {
		return nil, err
	}

	if dep.Kibana, err = NewKibanas(res.Resources.Kibana); err != nil {
		return nil, err
	}

	if diags := readApms(ctx, res.Resources.Apm, &dep.Apm); diags.HasError() {
		return nil, fmt.Errorf("diags: %v", diags)
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

func (dep *Deployment) Payload(client *api.API) (*models.DeploymentCreateRequest, error) {
	ctx := context.TODO()
	var result = models.DeploymentCreateRequest{
		Name:      dep.Name.Value,
		Alias:     dep.Alias.Value,
		Resources: &models.DeploymentCreateResources{},
		Settings:  &models.DeploymentCreateSettings{},
		Metadata:  &models.DeploymentCreateMetadata{},
	}

	dtID := dep.DeploymentTemplateId.Value
	version := dep.Version.Value

	template, err := deptemplateapi.Get(deptemplateapi.GetParams{
		API:                        client,
		TemplateID:                 dtID,
		Region:                     dep.Region.Value,
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

	esRes, err := dep.Elasticsearch.Payload(template, dtID, version, useNodeRoles)
	// enrichElasticsearchTemplate(
	// 	esResource(template), dtID, version, useNodeRoles,
	// ),

	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Elasticsearch = append(result.Resources.Elasticsearch, esRes...)

	kibanaRes, err := dep.Kibana.Payload(template)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.Kibana = append(result.Resources.Kibana, kibanaRes...)

	apms, diags := ApmPayload(ctx, template, dep.Apm)
	if diags.HasError() {
		merr = merr.Append(fmt.Errorf("diags: %v", diags))
	}
	result.Resources.Apm = append(result.Resources.Apm, apms...)

	integrationsServerRes, err := dep.IntegrationsServer.Payload(template)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.IntegrationsServer = append(result.Resources.IntegrationsServer, integrationsServerRes...)

	enterpriseSearchRes, err := dep.EnterpriseSearch.Payload(template)
	if err != nil {
		merr = merr.Append(err)
	}
	result.Resources.EnterpriseSearch = append(result.Resources.EnterpriseSearch, enterpriseSearchRes...)

	if err := merr.ErrorOrNil(); err != nil {
		return nil, err
	}

	trafficFilterToModel(dep.TrafficFilter, &result)

	observability, err := dep.Observability.Model(client)
	if err != nil {
		return nil, err
	}
	result.Settings.Observability = observability

	result.Metadata.Tags = converters.MapToTags(dep.Tags)

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
				dep.ElasticsearchUsername = types.String{Value: *creds.Username}
			}

			if creds.Password != nil && *creds.Password != "" {
				dep.ElasticsearchPassword = types.String{Value: *creds.Password}
			}
		}

		if res.SecretToken != "" {
			dep.ApmSecretToken = types.String{Value: res.SecretToken}
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
