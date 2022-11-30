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

package v2

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	v1 "github.com/elastic/terraform-provider-ec/ec/ecresource-tpf/deploymentresource/elasticsearch/v1"
)

type ElasticsearchExtensions v1.ElasticsearchExtensions

func ReadElasticsearchExtensions(in *models.ElasticsearchConfiguration) (ElasticsearchExtensions, error) {
	if len(in.UserBundles) == 0 && len(in.UserPlugins) == 0 {
		return nil, nil
	}

	extensions := make(ElasticsearchExtensions, 0, len(in.UserBundles)+len(in.UserPlugins))

	for _, model := range in.UserBundles {
		extension, err := v1.ReadFromUserBundle(model)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, *extension)
	}

	for _, model := range in.UserPlugins {
		extension, err := v1.ReadFromUserPlugin(model)
		if err != nil {
			return nil, err
		}

		extensions = append(extensions, *extension)
	}

	return extensions, nil
}
