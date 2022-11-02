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
    autoscale = "true"

    cold_tier = {
      size       = "0g"
      zone_count = 1
    }

    frozen_tier = {
      size       = "0g"
      zone_count = 1
    }

    hot_content_tier = {
      size       = "1g"
      zone_count = 1
      autoscaling = {
        max_size = "8g"
      }
    }

    ml_tier = {
      size       = "1g"
      zone_count = 1
      autoscaling = {
        min_size = "1g"
        max_size = "4g"
      }
    }

    warm_tier = {
      size       = "2g"
      zone_count = 1
      autoscaling = {
        max_size = "15g"
      }
    }
    
  }
}
