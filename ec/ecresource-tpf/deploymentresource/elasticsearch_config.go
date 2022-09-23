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
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchConfigs []*ElasticsearchConfig

func NewElasticsearchConfigs(in *models.ElasticsearchConfiguration) (ElasticsearchConfigs, error) {
	cfg, err := NewElasticsearchConfig(in)
	if err != nil {
		return nil, err
	}

	if !cfg.isEmpty() {
		return []*ElasticsearchConfig{cfg}, nil
	}

	return nil, nil
}

func (cfgs ElasticsearchConfigs) Payload(esCfg *models.ElasticsearchConfiguration) (*models.ElasticsearchConfiguration, error) {
	if len(cfgs) == 0 {
		return nil, nil
	}

	if esCfg == nil {
		esCfg = &models.ElasticsearchConfiguration{}
	}

	for _, cfg := range cfgs {
		if cfg.UserSettingsJson.Value != "" {
			if err := json.Unmarshal([]byte(cfg.UserSettingsJson.Value), &esCfg.UserSettingsJSON); err != nil {
				return nil, fmt.Errorf(
					"failed expanding elasticsearch user_settings_json: %w", err,
				)
			}
		}

		if cfg.UserSettingsOverrideJson.Value != "" {
			if err := json.Unmarshal([]byte(cfg.UserSettingsOverrideJson.Value), &esCfg.UserSettingsOverrideJSON); err != nil {
				return nil, fmt.Errorf(
					"failed expanding elasticsearch user_settings_override_json: %w", err,
				)
			}
		}

		if !cfg.UserSettingsYaml.IsNull() {
			esCfg.UserSettingsYaml = cfg.UserSettingsYaml.Value
		}

		if !cfg.UserSettingsOverrideYaml.IsNull() {
			esCfg.UserSettingsOverrideYaml = cfg.UserSettingsOverrideYaml.Value
		}

		if len(cfg.Plugins) > 0 {
			esCfg.EnabledBuiltInPlugins = cfg.Plugins
		}

		if !cfg.DockerImage.IsNull() {
			esCfg.DockerImage = cfg.DockerImage.Value
		}
	}

	return esCfg, nil
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

func NewElasticsearchConfig(in *models.ElasticsearchConfiguration) (*ElasticsearchConfig, error) {
	var cfg ElasticsearchConfig

	if in == nil {
		return nil, nil
	}

	if len(in.EnabledBuiltInPlugins) > 0 {
		cfg.Plugins = append(cfg.Plugins, in.EnabledBuiltInPlugins...)
	}

	if in.UserSettingsYaml != "" {
		cfg.UserSettingsYaml.Value = in.UserSettingsYaml
	}

	if in.UserSettingsOverrideYaml != "" {
		cfg.UserSettingsOverrideYaml.Value = in.UserSettingsOverrideYaml
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsJson.Value = string(b)
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			cfg.UserSettingsOverrideJson.Value = string(b)
		}
	}

	if in.DockerImage != "" {
		cfg.DockerImage.Value = in.DockerImage
	}

	return &cfg, nil
}
