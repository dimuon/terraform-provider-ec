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
	"sort"
	"strconv"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-ec/ec/internal/flatteners"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

type Elasticsearch struct {
	Autoscale     types.String            `tfsdk:"autoscale"`
	RefId         types.String            `tfsdk:"ref_id"`
	ResourceId    types.String            `tfsdk:"resource_id"`
	Region        types.String            `tfsdk:"region"`
	CloudID       types.String            `tfsdk:"cloud_id"`
	HttpEndpoint  types.String            `tfsdk:"http_endpoint"`
	HttpsEndpoint types.String            `tfsdk:"https_endpoint"`
	Topology      ElasticSearchTopologies `tfsdk:"topology"` //  []ElasticsearchTopology
	Config        ElasticsearchConfig     `tfsdk:"config"`
	// RemoteCluster  []ElasticsearchRemoteCluster `tfsdk:"remote_cluster"`
	// SnapshotSource ElasticsearchSnapshotSource  `tfsdk:"snapshot_source"`
	// Extension      []ElasticsearchExtension     `tfsdk:"extension"`
	// TrustAccount   []ElasticsearchTrustAccount  `tfsdk:"trust_account"`
	// TrustExternal  []ElasticsearchTrustExternal `tfsdk:"trust_external"`
	// Strategy       ElasticsearchStrategy        `tfsdk:"strategy"`
}

type ElasticSearchTopologies []ElasticsearchTopology

func (es *Elasticsearch) fromModel(in *models.ElasticsearchResourceInfo, name string, remotes models.RemoteResources) error {
	if util.IsCurrentEsPlanEmpty(in) || isEsResourceStopped(in) {
		return nil
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
	es.Topology.fromModel(plan.ClusterTopology, plan.AutoscalingEnabled != nil && *plan.AutoscalingEnabled)

	if plan.AutoscalingEnabled != nil {
		es.Autoscale.Value = strconv.FormatBool(*plan.AutoscalingEnabled)
	}

	if meta := in.Info.Metadata; meta != nil && meta.CloudID != "" {
		es.CloudID.Value = meta.CloudID
	}

	es.HttpEndpoint.Value, es.HttpsEndpoint.Value = flatteners.FlattenEndpoints(in.Info.Metadata)

	es.Config.fromModel(plan.Elasticsearch)

	// es.RemoteCluster.fromModel(remotes)

	// extensions := schema.NewSet(esExtensionHash, nil)
	// for _, ext := range flattenEsBundles(plan.Elasticsearch.UserBundles) {
	// 	extensions.Add(ext)
	// }

	// for _, ext := range flattenEsPlugins(plan.Elasticsearch.UserPlugins) {
	// 	extensions.Add(ext)
	// }

	// if extensions.Len() > 0 {
	// 	m["extension"] = extensions
	// }

	// if settings := res.Info.Settings; settings != nil {
	// 	if trust := flattenAccountTrust(settings.Trust); trust != nil {
	// 		m["trust_account"] = trust
	// 	}

	// 	if trust := flattenExternalTrust(settings.Trust); trust != nil {
	// 		m["trust_external"] = trust
	// 	}
	// }

	return nil
}

func (tops *ElasticSearchTopologies) fromModel(in []*models.ElasticsearchClusterTopologyElement, autoscaling bool) {
	if len(in) == 0 {
		return
	}
	if *tops == nil {
		*tops = make([]ElasticsearchTopology, 0, len(in))
	}
	for _, model := range in {
		if !isPotentiallySizedTopology(model, autoscaling) {
			continue
		}
		var top ElasticsearchTopology
		top.fromModel(model)
		*tops = append(*tops, top)
	}
	sort.SliceStable(*tops, func(i, j int) bool {
		a := (*tops)[i]
		b := (*tops)[j]
		return a.Id.Value < b.Id.Value
	})
}
