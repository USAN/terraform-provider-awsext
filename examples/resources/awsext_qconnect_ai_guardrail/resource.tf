resource "awsext_qconnect_ai_guardrail" "example" {
  assistant_id              = awsext_qconnect_assistant.example.assistant_id
  name                      = "example-guardrail"
  blocked_input_messaging   = "I can't help with that."
  blocked_outputs_messaging = "I can't share that."

  content_policy_config = {
    filters_config = [
      { input_strength = "HIGH", output_strength = "HIGH", type = "SEXUAL" },
      { input_strength = "HIGH", output_strength = "HIGH", type = "VIOLENCE" },
    ]
  }
}
