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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchConfigs types.List

func (configs ElasticsearchConfigs) Read(ctx context.Context, in *models.ElasticsearchConfiguration) diag.Diagnostics {
	var config ElasticsearchConfig

	diags := config.Read(in)

	if diags.HasError() {
		return diags
	}

	if config.isEmpty() {
		return nil
	}

	return tfsdk.ValueFrom(ctx, []*ElasticsearchConfig{&config}, elasticsearchTopology().FrameworkType(), configs)
}

func (configs ElasticsearchConfigs) Payload(ctx context.Context, model *models.ElasticsearchConfiguration) (*models.ElasticsearchConfiguration, diag.Diagnostics) {
	if len(configs.Elems) == 0 {
		return model, nil
	}

	if model == nil {
		model = &models.ElasticsearchConfiguration{}
	}

	for _, elem := range configs.Elems {
		var config ElasticsearchConfig

		diags := tfsdk.ValueAs(ctx, elem, &config)

		if diags.HasError() {
			return nil, diags
		}

		if config.UserSettingsJson.Value != "" {

			if err := json.Unmarshal([]byte(config.UserSettingsJson.Value), &model.UserSettingsJSON); err != nil {
				diags.AddError("failed expanding elasticsearch user_settings_json", err.Error())
				return nil, diags
			}
		}

		if config.UserSettingsOverrideJson.Value != "" {
			if err := json.Unmarshal([]byte(config.UserSettingsOverrideJson.Value), &model.UserSettingsOverrideJSON); err != nil {
				diags.AddError("failed expanding elasticsearch user_settings_override_json", err.Error())
				return nil, diags
			}
		}

		if !config.UserSettingsYaml.IsNull() {
			model.UserSettingsYaml = config.UserSettingsYaml.Value
		}

		if !config.UserSettingsOverrideYaml.IsNull() {
			model.UserSettingsOverrideYaml = config.UserSettingsOverrideYaml.Value
		}

		if len(config.Plugins) > 0 {
			model.EnabledBuiltInPlugins = config.Plugins
		}

		if !config.DockerImage.IsNull() {
			model.DockerImage = config.DockerImage.Value
		}
	}

	return model, nil
}

type ElasticsearchConfig struct {
	Plugins                  []string     `tfsdk:"plugins"`
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

func (c *ElasticsearchConfig) isEmpty() bool {
	return reflect.ValueOf(*c).IsZero()
}

func (config *ElasticsearchConfig) Read(in *models.ElasticsearchConfiguration) diag.Diagnostics {
	if in == nil {
		return nil
	}

	if len(in.EnabledBuiltInPlugins) > 0 {
		config.Plugins = append(config.Plugins, in.EnabledBuiltInPlugins...)
	}

	if in.UserSettingsYaml != "" {
		config.UserSettingsYaml = types.String{Value: in.UserSettingsYaml}
	}

	if in.UserSettingsOverrideYaml != "" {
		config.UserSettingsOverrideYaml = types.String{Value: in.UserSettingsOverrideYaml}
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			config.UserSettingsJson = types.String{Value: string(b)}
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			config.UserSettingsOverrideJson = types.String{Value: string(b)}
		}
	}

	if in.DockerImage != "" {
		config.DockerImage = types.String{Value: in.DockerImage}
	}

	return nil
}
