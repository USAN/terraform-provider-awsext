resource "awsext_qconnect_assistant" "example" {
  name        = "example-assistant"
  type        = "AGENT"
  description = "Example Q in Connect assistant"
  tags        = { env = "dev" }
}
