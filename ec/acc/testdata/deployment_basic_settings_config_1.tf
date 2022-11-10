data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "basic" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {

    config = {}

    hot = {
      size        = "1g"
      autoscaling = {}
    }
  }

  kibana = {
    topology = {
      instance_configuration_id = "%s"
    }
  }

  apm = {
    topology = {
      instance_configuration_id = "%s"
    }
  }

  enterprise_search = {
    topology = {
      instance_configuration_id = "%s"
    }
  }
}