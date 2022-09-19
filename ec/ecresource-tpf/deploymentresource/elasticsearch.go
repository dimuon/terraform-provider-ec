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
	"strconv"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/flatteners"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type Elasticsearch struct {
	Autoscale      types.String                  `tfsdk:"autoscale"`
	RefId          types.String                  `tfsdk:"ref_id"`
	ResourceId     types.String                  `tfsdk:"resource_id"`
	Region         types.String                  `tfsdk:"region"`
	CloudID        types.String                  `tfsdk:"cloud_id"`
	HttpEndpoint   types.String                  `tfsdk:"http_endpoint"`
	HttpsEndpoint  types.String                  `tfsdk:"https_endpoint"`
	Topology       []ElasticsearchTopology       `tfsdk:"topology"`
	Config         []ElasticsearchConfig         `tfsdk:"config"`
	RemoteCluster  []ElasticsearchRemoteCluster  `tfsdk:"remote_cluster"`
	SnapshotSource []ElasticsearchSnapshotSource `tfsdk:"snapshot_source"`
	Extension      []ElasticsearchExtension      `tfsdk:"extension"`
	TrustAccount   []ElasticsearchTrustAccount   `tfsdk:"trust_account"`
	TrustExternal  []ElasticsearchTrustExternal  `tfsdk:"trust_external"`
	Strategy       []ElasticsearchStrategy       `tfsdk:"strategy"`
}

func NewElasticsearches(in []*models.ElasticsearchResourceInfo, remotes *models.RemoteResources) ([]Elasticsearch, error) {
	if len(in) == 0 {
		return nil, nil
	}

	ess := make([]Elasticsearch, 0, len(in))

	for _, model := range in {
		if util.IsCurrentEsPlanEmpty(model) || isEsResourceStopped(model) {
			continue
		}
		var es *Elasticsearch
		var err error
		if es, err = NewElasticsearch(model, remotes); err != nil {
			return nil, err
		}
		ess = append(ess, *es)
	}

	return ess, nil
}

func NewElasticsearch(in *models.ElasticsearchResourceInfo, remotes *models.RemoteResources) (*Elasticsearch, error) {
	var es Elasticsearch

	if util.IsCurrentEsPlanEmpty(in) || isEsResourceStopped(in) {
		return &es, nil
	}

	if in.Info.ClusterID != nil && *in.Info.ClusterID != "" {
		es.ResourceId.Value = *in.Info.ClusterID
	}

	if in.RefID != nil && *in.RefID != "" {
		es.RefId.Value = *in.RefID
	}

	if in.Region != nil {
		es.Region.Value = *in.Region
	}

	plan := in.Info.PlanInfo.Current.Plan
	var err error

	es.Topology, err = NewTopologies(plan.ClusterTopology, plan.AutoscalingEnabled != nil && *plan.AutoscalingEnabled)
	if err != nil {
		return &es, err
	}

	if plan.AutoscalingEnabled != nil {
		es.Autoscale.Value = strconv.FormatBool(*plan.AutoscalingEnabled)
	}

	if meta := in.Info.Metadata; meta != nil && meta.CloudID != "" {
		es.CloudID.Value = meta.CloudID
	}

	es.HttpEndpoint.Value, es.HttpsEndpoint.Value = flatteners.FlattenEndpoints(in.Info.Metadata)

	es.Config, err = NewElasticsearchConfigs(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}

	es.RemoteCluster, err = NewElasticsearchRemoteClusters(remotes.Resources)
	if err != nil {
		return nil, err
	}

	es.Extension, err = NewElasticsearchExtensions(plan.Elasticsearch)
	if err != nil {
		return nil, err
	}

	es.TrustAccount, err = NewElasticsearchTrustAccounts(in.Info.Settings.Trust)
	if err != nil {
		return nil, err
	}

	es.TrustExternal, err = NewElasticsearchTrustExternals(in.Info.Settings.Trust)
	if err != nil {
		return nil, err
	}

	return &es, nil
}
