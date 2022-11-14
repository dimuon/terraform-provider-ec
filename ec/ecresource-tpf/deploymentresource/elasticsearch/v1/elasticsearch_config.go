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
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchConfigTF struct {
	Plugins                  types.Set    `tfsdk:"plugins"`
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

type ElasticsearchConfigs []ElasticsearchConfig

func ElasticsearchConfigsPayload(ctx context.Context, cfgList types.List, model *models.ElasticsearchConfiguration) (*models.ElasticsearchConfiguration, diag.Diagnostics) {
	if cfgList.IsNull() || cfgList.IsUnknown() || len(cfgList.Elems) == 0 {
		return model, nil
	}

	return ElasticsearchConfigPayload(ctx, cfgList.Elems[0], model)
}

func ElasticsearchConfigPayload(ctx context.Context, cfgObj attr.Value, model *models.ElasticsearchConfiguration) (*models.ElasticsearchConfiguration, diag.Diagnostics) {
	if cfgObj.IsNull() || cfgObj.IsUnknown() {
		return model, nil
	}

	var cfg ElasticsearchConfigTF

	diags := tfsdk.ValueAs(ctx, cfgObj, &cfg)

	if diags.HasError() {
		return nil, diags
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

	if !cfg.DockerImage.IsNull() {
		model.DockerImage = cfg.DockerImage.Value
	}

	return model, diags
}

func ReadElasticsearchConfigs(in *models.ElasticsearchConfiguration) (ElasticsearchConfigs, error) {
	config, err := ReadElasticsearchConfig(in)

	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return ElasticsearchConfigs{*config}, nil
}

func ReadElasticsearchConfig(in *models.ElasticsearchConfiguration) (*ElasticsearchConfig, error) {
	var config ElasticsearchConfig

	if in == nil {
		return &ElasticsearchConfig{}, nil
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

	return &config, nil
}
