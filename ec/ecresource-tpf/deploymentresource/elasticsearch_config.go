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

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ElasticsearchConfig struct {
	DockerImage              types.String `tfsdk:"docker_image"`
	Plugins                  types.Set    `tfsdk:"plugins"`
	UserSettingsJson         types.String `tfsdk:"user_settings_json"`
	UserSettingsOverrideJson types.String `tfsdk:"user_settings_override_json"`
	UserSettingsYaml         types.String `tfsdk:"user_settings_yaml"`
	UserSettingsOverrideYaml types.String `tfsdk:"user_settings_override_yaml"`
}

func (esc *ElasticsearchConfig) fromModel(in *models.ElasticsearchConfiguration) error {
	if in == nil {
		return nil
	}

	if len(in.EnabledBuiltInPlugins) > 0 {
		esc.Plugins.ElemType = types.StringType
		esc.Plugins.Elems = make([]attr.Value, 0, len(in.EnabledBuiltInPlugins))
		for _, plugin := range in.EnabledBuiltInPlugins {
			esc.Plugins.Elems = append(esc.Plugins.Elems, types.String{Value: plugin})
		}
	}

	if in.UserSettingsYaml != "" {
		esc.UserSettingsYaml.Value = in.UserSettingsYaml
	}

	if in.UserSettingsOverrideYaml != "" {
		esc.UserSettingsOverrideYaml.Value = in.UserSettingsOverrideYaml
	}

	if o := in.UserSettingsJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			esc.UserSettingsJson.Value = string(b)
		}
	}

	if o := in.UserSettingsOverrideJSON; o != nil {
		if b, _ := json.Marshal(o); len(b) > 0 && !bytes.Equal([]byte("{}"), b) {
			esc.UserSettingsOverrideJson.Value = string(b)

		}
	}

	if in.DockerImage != "" {
		esc.DockerImage.Value = in.DockerImage
	}

	return nil
}
