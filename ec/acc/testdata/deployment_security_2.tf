data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "security" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    config = {}
    hot = {
      size        = "2g"
      autoscaling = {}
    }
  }

  kibana = { topology = {} }

  apm = { topology = {} }
}