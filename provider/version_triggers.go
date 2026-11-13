package provider

// triggersDescription documents the `triggers` attribute shared by the
// qconnect version resources.
//
// The value is stored in state only after the version was created, so a failed
// create leaves the previous value in state and the next plan still shows the
// difference. That is the property a separate terraform_data hash cannot give:
// terraform_data saves its hash before the version create runs.
const triggersDescription = "Arbitrary map of values that, when changed, forces a new version to be published. " +
	"Typically a hash of the content being versioned. The value is saved to state only after the " +
	"version was created successfully, so a failed create is retried on the next apply."
