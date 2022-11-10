data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "defaults" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    config = {}
    hot = {
      autoscaling = {}
    }
  }

  kibana = {
    topology = {}
  }

  enterprise_search = {
    topology = {
      zone_count = 1
    }
  }
}