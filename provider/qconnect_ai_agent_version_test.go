package provider

import "testing"

func TestVersionedAIAgentArn(t *testing.T) {
	cases := []struct {
		name          string
		agentArn      string
		versionNumber int64
		want          string
	}{
		{
			name:          "bare entity ARN from CreateAIAgentVersion gets the suffix appended",
			agentArn:      "arn:aws:wisdom:us-east-1:311923416029:ai-agent/e363ea07-690c-4459-aa72-ea588b391bdc/357bc310-2dff-4760-b397-aa203336534c",
			versionNumber: 1,
			want:          "arn:aws:wisdom:us-east-1:311923416029:ai-agent/e363ea07-690c-4459-aa72-ea588b391bdc/357bc310-2dff-4760-b397-aa203336534c:1",
		},
		{
			name:          "already-versioned ARN from GetAIAgent is not double-suffixed",
			agentArn:      "arn:aws:wisdom:us-east-1:311923416029:ai-agent/e363ea07-690c-4459-aa72-ea588b391bdc/357bc310-2dff-4760-b397-aa203336534c:1",
			versionNumber: 1,
			want:          "arn:aws:wisdom:us-east-1:311923416029:ai-agent/e363ea07-690c-4459-aa72-ea588b391bdc/357bc310-2dff-4760-b397-aa203336534c:1",
		},
		{
			name:          "higher version number",
			agentArn:      "arn:aws:wisdom:us-east-1:311923416029:ai-agent/e363ea07-690c-4459-aa72-ea588b391bdc/357bc310-2dff-4760-b397-aa203336534c:7",
			versionNumber: 7,
			want:          "arn:aws:wisdom:us-east-1:311923416029:ai-agent/e363ea07-690c-4459-aa72-ea588b391bdc/357bc310-2dff-4760-b397-aa203336534c:7",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := versionedAIAgentArn(c.agentArn, c.versionNumber)
			if got != c.want {
				t.Errorf("versionedAIAgentArn(%q, %d) = %q, want %q", c.agentArn, c.versionNumber, got, c.want)
			}
		})
	}
}
