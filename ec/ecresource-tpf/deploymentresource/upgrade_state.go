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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	apmv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/apm/v2"
	deploymentv1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/deployment/v1"
	deploymentv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/deployment/v2"
	elasticsearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/elasticsearch/v2"
	enterprisesearchv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/enterprisesearch/v2"
	integrationsserverv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/integrationsserver/v2"
	kibanav2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/kibana/v2"
	observabilityv2 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/observability/v2"
)

func (r *Resource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := deploymentv1.DeploymentSchema()
	return map[int64]resource.StateUpgrader{
		1: {
			PriorSchema: &schemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				UpgradeStateToV2(ctx, req, resp)
			},
		},
	}
}

func UpgradeStateToV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var priorStateData deploymentv1.DeploymentTF

	resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)

	upgradedStateData := deploymentv2.DeploymentTF{
		Id:                    priorStateData.Id,
		Alias:                 priorStateData.Alias,
		Version:               priorStateData.Version,
		Region:                priorStateData.Region,
		DeploymentTemplateId:  priorStateData.DeploymentTemplateId,
		Name:                  priorStateData.Name,
		RequestId:             priorStateData.RequestId,
		ElasticsearchUsername: priorStateData.ElasticsearchUsername,
		ElasticsearchPassword: priorStateData.ElasticsearchPassword,
		ApmSecretToken:        priorStateData.ApmSecretToken,
		TrafficFilter:         priorStateData.TrafficFilter,
		Tags:                  priorStateData.Tags,
		Elasticsearch:         ElasticsearchV1ToV2(priorStateData.Elasticsearch),
		//Elasticsearch:         types.Object{Null: true, AttrTypes: elasticsearchv2.ElasticsearchSchema().Attributes.Type().(types.ObjectType).AttrTypes},
		Kibana:             types.Object{Null: true, AttrTypes: kibanav2.KibanaSchema().Attributes.Type().(types.ObjectType).AttrTypes},
		Apm:                types.Object{Null: true, AttrTypes: apmv2.ApmSchema().Attributes.Type().(types.ObjectType).AttrTypes},
		IntegrationsServer: types.Object{Null: true, AttrTypes: integrationsserverv2.IntegrationsServerSchema().Attributes.Type().(types.ObjectType).AttrTypes},
		EnterpriseSearch:   types.Object{Null: true, AttrTypes: enterprisesearchv2.EnterpriseSearchSchema().Attributes.Type().(types.ObjectType).AttrTypes},
		Observability:      types.Object{Null: true, AttrTypes: observabilityv2.ObservabilitySchema().Attributes.Type().(types.ObjectType).AttrTypes},
		//Elasticsearch:         ElasticsearchV1ToElasticsearchV2(priorStateData.Elasticsearch),
		//Kibana:                KibanaV1ToKibanaV2(priorStateData.Kibana),
		//Apm: priorStateData.Apm,
		//IntegrationsServer: priorStateData.IntegrationsServer,
		//EnterpriseSearch: priorStateData.EnterpriseSearch,
		//Observability: priorStateData.Observability,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
}

func ElasticsearchV1ToV2(list types.List) types.Object {
	//fmt.Println(dd.Dump(list, dd.WithIndent(4)))

	if list.IsNull() || len(list.Elems) == 0 || list.Elems[0].IsNull() {
		return types.Object{Null: true}
	}
	elasticsearch := list.Elems[0].(types.Object)
	elasticsearch.AttrTypes = elasticsearchv2.ElasticsearchSchema().Attributes.Type().(types.ObjectType).AttrTypes

	if value, ok := elasticsearch.Attrs["topology"]; ok {
		topologyList := value.(types.List)

		hot := types.Object{Null: true, AttrTypes: elasticsearchv2.HotTierSchema.Attributes.Type().(types.ObjectType).AttrTypes}
		coordinating := types.Object{Null: true, AttrTypes: elasticsearchv2.CoordinatingTierSchema.Attributes.Type().(types.ObjectType).AttrTypes}
		master := types.Object{Null: true, AttrTypes: elasticsearchv2.MasterTierSchema.Attributes.Type().(types.ObjectType).AttrTypes}
		warm := types.Object{Null: true, AttrTypes: elasticsearchv2.WarmTierSchema.Attributes.Type().(types.ObjectType).AttrTypes}
		cold := types.Object{Null: true, AttrTypes: elasticsearchv2.ColdTierSchema.Attributes.Type().(types.ObjectType).AttrTypes}
		frozen := types.Object{Null: true, AttrTypes: elasticsearchv2.FrozenTierSchema.Attributes.Type().(types.ObjectType).AttrTypes}
		ml := types.Object{Null: true, AttrTypes: elasticsearchv2.MlTierSchema.Attributes.Type().(types.ObjectType).AttrTypes}

		for _, elem := range topologyList.Elems {
			topology := elem.(types.Object)
			if !topology.Null && !topology.Attrs["id"].(types.String).Null {
				id := topology.Attrs["id"].(types.String).Value
				convertedTopology := ElasticsearchTopologyV1ToV2(topology)
				switch id {
				case "hot_content":
					convertedTopology.AttrTypes = hot.AttrTypes
					hot = convertedTopology
				case "warm":
					convertedTopology.AttrTypes = warm.AttrTypes
					warm = convertedTopology
				case "cold":
					convertedTopology.AttrTypes = cold.AttrTypes
					cold = convertedTopology
				case "coordinating":
					convertedTopology.AttrTypes = coordinating.AttrTypes
					coordinating = convertedTopology
				case "frozen":
					convertedTopology.AttrTypes = frozen.AttrTypes
					frozen = convertedTopology
				case "ml":
					convertedTopology.AttrTypes = ml.AttrTypes
					ml = convertedTopology
				case "master":
					convertedTopology.AttrTypes = master.AttrTypes
					master = convertedTopology
				}
			}
		}
		// TODO

		elasticsearch.Attrs["hot"] = hot
		elasticsearch.Attrs["coordinating"] = coordinating
		elasticsearch.Attrs["master"] = master
		elasticsearch.Attrs["warm"] = warm
		elasticsearch.Attrs["cold"] = cold
		elasticsearch.Attrs["frozen"] = frozen
		elasticsearch.Attrs["ml"] = ml
		delete(elasticsearch.Attrs, "topology")
	}

	// config was a list and is an object now. Schema stayed the same apart from that.
	if value, ok := elasticsearch.Attrs["config"]; ok {
		config := types.Object{Null: true, AttrTypes: elasticsearchv2.ElasticsearchConfigSchema().Attributes.Type().(types.ObjectType).AttrTypes}
		configList := value.(types.List)
		if len(configList.Elems) > 0 {
			config = configList.Elems[0].(types.Object)
		}
		elasticsearch.Attrs["config"] = config
	}

	// snapshot_source was a list and is an object now. Schema stayed the same apart from that.
	if value, ok := elasticsearch.Attrs["snapshot_source"]; ok {
		snapshotSource := types.Object{Null: true, AttrTypes: elasticsearchv2.ElasticsearchSnapshotSourceSchema().Attributes.Type().(types.ObjectType).AttrTypes}
		snapshotSourceList := value.(types.List)
		if len(snapshotSourceList.Elems) > 0 {
			snapshotSource = snapshotSourceList.Elems[0].(types.Object)
		}
		elasticsearch.Attrs["snapshot_source"] = snapshotSource
	}

	// strategy used to be a list of structs with a field "type". This is just a string now.
	if _, ok := elasticsearch.Attrs["strategy"]; ok {
		elasticsearch.Attrs["strategy"] = ElasticsearchStrategyV1ToV2(elasticsearch.Attrs["strategy"])

	}
	//fmt.Println(dd.Dump(elasticsearch, dd.WithIndent(4)))

	return elasticsearch
}

func ElasticsearchStrategyV1ToV2(old attr.Value) types.String {
	list := old.(types.List)
	strategy := types.String{Null: true}
	if len(list.Elems) > 0 {
		object := list.Elems[0].(types.Object)
		if !object.Null {
			if value, ok := object.Attrs["type"]; ok {
				strategy = value.(types.String)
			}
		}
	}
	return strategy
}

func ElasticsearchTopologyV1ToV2(topology types.Object) types.Object {
	if topology.Null {
		topology.Attrs = map[string]attr.Value{}
		return topology
	}

	// id attribute has been removed
	delete(topology.Attrs, "id")

	// autoscaling was a list and is an object now. Schema stayed the same apart from that.
	if value, ok := topology.Attrs["autoscaling"]; ok {
		autoscaling := types.Object{Null: true, AttrTypes: elasticsearchv2.ElasticsearchTopologyAutoscalingSchema("hot").Attributes.Type().(types.ObjectType).AttrTypes}
		autoscalingList := value.(types.List)
		if len(autoscalingList.Elems) > 0 {
			autoscaling = autoscalingList.Elems[0].(types.Object)
		}
		topology.Attrs["autoscaling"] = autoscaling
	}

	return topology
}

/*

func KibanaV1ToKibanaV2(kibanas kibanav1.Kibanas) *kibanav2.Kibana {
	if len(kibanas) == 0 {
		return nil
	}

	kibana := kibanas[0]
	return &kibanav2.Kibana{
		ElasticsearchClusterRefId: kibana.ElasticsearchClusterRefId,
		RefId:                     kibana.RefId,
		ResourceId:                kibana.ResourceId,
		Region:                    kibana.Region,
		HttpEndpoint:              kibana.HttpEndpoint,
		HttpsEndpoint:             kibana.HttpsEndpoint,
		Topology:                  KibanaTopologyV1ToKibanaTopologyV2(kibana.Topology),
		Config:                    KibanaConfigV1ToKibanaConfigV2(kibana.Config),
	}
}

func KibanaTopologyV1ToKibanaTopologyV2(topologies topologyv1.Topologies) *topologyv1.Topology {
	if len(topologies) == 0 {
		return nil
	}
	return &topologies[0]
}

func KibanaConfigV1ToKibanaConfigV2(configs kibanav1.KibanaConfigs) *kibanav1.KibanaConfig {
	if len(configs) == 0 {
		return nil
	}
	return &configs[0]
}
*/
