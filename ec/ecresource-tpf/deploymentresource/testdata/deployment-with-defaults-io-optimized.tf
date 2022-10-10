resource "ec_deployment" "empty declarations (IO Optimized)" {
    name = "my_deployment_name"
    deployment_template_id = "aws-io-optimized-v2"
    region = "us-east-1"
    version = "7.7.0"
    traffic_filter = ["0.0.0.0/0", "192.168.10.0/24"]
}