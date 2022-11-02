data "ec_stack" "autoscaling" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "autoscaling" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.autoscaling.version
  deployment_template_id = "%s"

  elasticsearch = {
    autoscale = "false"

    cold_tier = {
      size       = "0g"
      zone_count = 1
    },

    frozne_tier = {
      size       = "0g"
      zone_count = 1
    },

    hot_content_tier = {
      id         = "hot_content"
      size       = "1g"
      zone_count = 1
      autoscaling = {
        max_size = "8g"
      }
    },

    ml_tier = {
      size       = "0g"
      zone_count = 1
      autoscaling = {
        min_size = "0g"
        max_size = "4g"
      }
    },

    warm_tier = {
      size       = "2g"
      zone_count = 1
      autoscaling = {
        max_size = "15g"
      }
    }
  }
}
