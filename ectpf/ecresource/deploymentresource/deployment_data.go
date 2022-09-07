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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeploymentData struct {
	Version              types.String      `tfsdk:"version"`
	Region               types.String      `tfsdk:"region"`
	DeploymentTemplateId types.String      `tfsdk:"deployment_template_id"`
	Name                 types.String      `tfsdk:"name"`
	Elasticsearch        ElasticsearchData `tfsdk:"elasticsearch"`
}

type ElasticsearchData struct {
	Autoscale     types.String                `tfsdk:"autoscale"`
	ResourceId    types.String                `tfsdk:"resource_id"`
	Region        types.String                `tfsdk:"region"`
	HttpsEndpoint types.String                `tfsdk:"https_endpoint"`
	Topology      []ElasticsearchTopologyData `tfsdk:"topology"`
}

type ElasticsearchTopologyData struct {
	Id types.String `tfsdk:"id"`
}
