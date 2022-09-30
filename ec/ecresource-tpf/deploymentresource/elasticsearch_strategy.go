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

type ElasticsearchStrategies []ElasticsearchStrategy

func (strategies ElasticsearchStrategiesTF) Payload(ctx context.Context, model *models.TransientElasticsearchPlanConfiguration) (*models.TransientElasticsearchPlanConfiguration, diag.Diagnostics) {
	if len(strategies.Elems) == 0 {
		return nil, nil
	}

	if model == nil {
		model = &models.TransientElasticsearchPlanConfiguration{
			Strategy: &models.PlanStrategy{},
		}
	}

	for _, elem := range strategies.Elems {
		var strategy ElasticsearchStrategyTF
		if diags := tfsdk.ValueAs(ctx, elem, &strategy); diags.HasError() {
			return nil, diags
		}
		switch strategy.Type.Value {
		case autodetect:
			model.Strategy.Autodetect = new(models.AutodetectStrategyConfig)
		case growAndShrink:
			model.Strategy.GrowAndShrink = new(models.GrowShrinkStrategyConfig)
		case rollingGrowAndShrink:
			model.Strategy.RollingGrowAndShrink = new(models.RollingGrowShrinkStrategyConfig)
		case rollingAll:
			model.Strategy.Rolling = &models.RollingStrategyConfig{
				GroupBy: "__all__",
			}
		}
	}

	return model, nil
}
