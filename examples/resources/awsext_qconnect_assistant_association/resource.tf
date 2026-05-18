resource "awsext_qconnect_assistant_association" "bedrock" {
  assistant_id     = awsext_qconnect_assistant.example.assistant_id
  association_type = "EXTERNAL_BEDROCK_KNOWLEDGE_BASE"

  association = {
    bedrock_knowledge_base = {
      knowledge_base_arn = "arn:aws:bedrock:us-east-1::knowledge-base/ABCDEF1234"
      access_role_arn    = "arn:aws:iam::111111111111:role/QConnectKBRole"
    }
  }
}

resource "awsext_qconnect_assistant_association" "native_kb" {
  assistant_id     = awsext_qconnect_assistant.example.assistant_id
  association_type = "KNOWLEDGE_BASE"

  association = {
    knowledge_base_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  }
}
