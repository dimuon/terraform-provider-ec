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
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewElasticsearchExtensions(in *models.ElasticsearchConfiguration) ([]ElasticsearchExtension, error) {
	if len(in.UserBundles) == 0 && len(in.UserPlugins) == 0 {
		return nil, nil
	}

	exts := make([]ElasticsearchExtension, 0, len(in.UserBundles)+len(in.UserPlugins))

	for _, model := range in.UserBundles {
		ext, err := NewElasticsearchExtensionFromUserBundle(model)

		if err != nil {
			return nil, err
		}
		exts = append(exts, *ext)
	}

	for _, model := range in.UserPlugins {
		ext, err := ElasticsearchExtensionFromUserPlugin(model)
		if err != nil {
			return nil, err
		}
		exts = append(exts, *ext)
	}

	return exts, nil
}

type ElasticsearchExtension struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
	Url     types.String `tfsdk:"url"`
}

func NewElasticsearchExtensionFromUserBundle(in *models.ElasticsearchUserBundle) (*ElasticsearchExtension, error) {
	var ext ElasticsearchExtension

	ext.Type.Value = "bundle"

	if in.ElasticsearchVersion == nil {
		return nil, missingField("ElasticsearchUserBundle.ElasticsearchVersion")
	}
	ext.Version.Value = *in.ElasticsearchVersion

	if in.URL == nil {
		return nil, missingField("ElasticsearchUserBundle.URL")
	}
	ext.Url.Value = *in.URL

	if in.Name == nil {
		return nil, missingField("ElasticsearchUserBundle.Name")
	}
	ext.Name.Value = *in.Name

	return &ext, nil
}

func ElasticsearchExtensionFromUserPlugin(in *models.ElasticsearchUserPlugin) (*ElasticsearchExtension, error) {
	var ext ElasticsearchExtension

	ext.Type.Value = "plugin"

	if in.ElasticsearchVersion == nil {
		return nil, missingField("ElasticsearchUserPlugin.ElasticsearchVersion")
	}
	ext.Version.Value = *in.ElasticsearchVersion

	if in.URL == nil {
		return nil, missingField("ElasticsearchUserPlugin.URL")
	}
	ext.Url.Value = *in.URL

	if in.Name == nil {
		return nil, missingField("ElasticsearchUserPlugin.Name")
	}
	ext.Name.Value = *in.Name

	return &ext, nil
}
