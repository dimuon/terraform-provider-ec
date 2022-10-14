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
	"bytes"
	"context"
	"encoding/json"
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchTopologyConfigTF struct {
	Plugins                  types.Set    `tfsdk:"plugins"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ElasticsearchTopologyConfig struct {
	Plugins                  []string `tfsdk:"plugins"`
	UserSettingsJson         *string  `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson *string  `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         *string  `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml *string  `tfsdk:"user_settings_override_yaml"`
}

type ElasticsearchTopologyConfigs []ElasticsearchTopologyConfig

func elasticsearchTopologyConfigPayload(ctx context.Context, list types.List, model *models.ElasticsearchConfiguration) (*models.ElasticsearchConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	var cfg *ElasticsearchTopologyConfigTF

	if diags = getFirst(ctx, list, &cfg); diags.HasError() {
		return nil, diags
	}

	if cfg == nil {
		return model, nil
	}

	if model == nil {
		model = &models.ElasticsearchConfiguration{}
	}

	if cfg.UserSettingsJson.Value != "" {

		if err := json.Unmarshal([]byte(cfg.UserSettingsJson.Value), &model.UserSettingsJSON); err != nil {
			diags.AddError("failed expanding elasticsearch user_settings_json", err.Error())
		}
	}

	if cfg.UserSettingsOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.Value), &model.UserSettingsOverrideJSON); err != nil {
			diags.AddError("failed expanding elasticsearch user_settings_override_json", err.Error())
		}
	}

	if !cfg.UserSettingsYaml.IsNull() {
		model.UserSettingsYaml = cfg.UserSettingsYaml.Value
	}

	if !cfg.UserSettingsOverrideYaml.IsNull() {
		model.UserSettingsOverrideYaml = cfg.UserSettingsOverrideYaml.Value
	}

	ds := cfg.Plugins.ElementsAs(ctx, &model.EnabledBuiltInPlugins, true)

	diags = append(diags, ds...)

	return model, diags
}

func (c *ElasticsearchTopologyConfig) isEmpty() bool {
	return c == nil || reflect.ValueOf(*c).IsZero()
}

func readElasticsearchTopologyConfig(in *models.ElasticsearchConfiguration) (ElasticsearchTopologyConfigs, error) {
	var config ElasticsearchTopologyConfig
	if in == nil {
		return nil, nil
	}

	if len(in.EnabledBuiltInPlugins) > 0 {
		config.Plugins = append(config.Plugins, in.EnabledBuiltInPlugins...)
	}

	if in.UserSettingsYaml != "" {
		config.UserSettingsYaml = &in.UserSettingsYaml
	}

	if in.UserSettingsOverrideYaml != "" {
		config.UserSettingsOverrideYaml = &in.UserSettingsOverrideYaml
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			config.UserSettingsJson = ec.String(string(b))
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			config.UserSettingsOverrideJson = ec.String(string(b))
		}
	}

	if config.isEmpty() {
		return nil, nil
	}

	return ElasticsearchTopologyConfigs{config}, nil
}
