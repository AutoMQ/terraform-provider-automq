resource "automq_connector_plugin" "s3_sink" {
  environment_id  = "env-xxxxx"
  name            = "s3-sink-connector"
  version         = "10.5.17"
  storage_url     = "s3://my-bucket/plugins/confluentinc-kafka-connect-s3-10.5.17.zip"
  types           = ["SINK"]
  connector_class = "io.confluent.connect.s3.S3SinkConnector"
  description     = "Confluent S3 Sink Connector"
}
