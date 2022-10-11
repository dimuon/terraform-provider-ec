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

type ElasticsearchConfigTF struct {
	Plugins                  []string     `tfsdk:"plugins"`
	DockerImage              types.String `tfsdk:"docker_image"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

type ElasticsearchConfig struct {
	Plugins                  []string `tfsdk:"plugins"`
	DockerImage              *string  `tfsdk:"docker_image"`
	UserSettingsJson         *string  `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson *string  `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         *string  `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml *string  `tfsdk:"user_settings_override_yaml"`
}

func (config *ElasticsearchConfigTF) Payload(ctx context.Context, model *models.ElasticsearchConfiguration) (*models.ElasticsearchConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if config == nil {
		return model, nil
	}

	if model == nil {
		model = &models.ElasticsearchConfiguration{}
	}

	if config.UserSettingsJson.Value != "" {

		if err := json.Unmarshal([]byte(config.UserSettingsJson.Value), &model.UserSettingsJSON); err != nil {
			diags.AddError("failed expanding elasticsearch user_settings_json", err.Error())
		}
	}

	if config.UserSettingsOverrideJson.Value != "" {
		if err := json.Unmarshal([]byte(config.UserSettingsOverrideJson.Value), &model.UserSettingsOverrideJSON); err != nil {
			diags.AddError("failed expanding elasticsearch user_settings_override_json", err.Error())
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

	return model, diags
}

func (c *ElasticsearchConfig) isEmpty() bool {
	return c == nil || reflect.ValueOf(*c).IsZero()
}

func readElasticsearchConfig(in *models.ElasticsearchConfiguration) (*ElasticsearchConfig, error) {
	var config ElasticsearchConfig
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

	if in.DockerImage != "" {
		config.DockerImage = ec.String(in.DockerImage)
	}

	if config.isEmpty() {
		return nil, nil
	}

	return &config, nil
}
