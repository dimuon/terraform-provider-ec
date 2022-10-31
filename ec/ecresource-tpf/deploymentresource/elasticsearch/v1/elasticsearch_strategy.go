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

package v1

import (
	"context"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchStrategyTF struct {
	Type types.String `tfsdk:"type"`
}

type ElasticsearchStrategiesTF types.List

type ElasticsearchStrategy struct {
	Type string `tfsdk:"type"`
}

type ElasticsearchStrategiesV1 []ElasticsearchStrategy

func ElasticsearchStrategiesPayload(ctx context.Context, strategies types.List, payload *models.ElasticsearchClusterPlan) diag.Diagnostics {
	if len(strategies.Elems) == 0 {
		return nil
	}

	if payload.Transient == nil {
		payload.Transient = &models.TransientElasticsearchPlanConfiguration{
			Strategy: &models.PlanStrategy{},
		}
	}

	for _, elem := range strategies.Elems {
		var strategy ElasticsearchStrategyTF
		if diags := tfsdk.ValueAs(ctx, elem, &strategy); diags.HasError() {
			return diags
		}
		ElasticsearchStrategyPayload(strategy.Type, payload)
	}

	return nil
}

func ElasticsearchStrategyPayload(strategy types.String, payload *models.ElasticsearchClusterPlan) {
	switch strategy.Value {
	case autodetect:
		payload.Transient.Strategy.Autodetect = new(models.AutodetectStrategyConfig)
	case growAndShrink:
		payload.Transient.Strategy.GrowAndShrink = new(models.GrowShrinkStrategyConfig)
	case rollingGrowAndShrink:
		payload.Transient.Strategy.RollingGrowAndShrink = new(models.RollingGrowShrinkStrategyConfig)
	case rollingAll:
		payload.Transient.Strategy.Rolling = &models.RollingStrategyConfig{
			GroupBy: "__all__",
		}
	}
}
