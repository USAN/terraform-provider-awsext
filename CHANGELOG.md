# Changelog

All notable changes to this provider are documented here. Versions correspond to git tags.

## v1.6.2

### Fixed

- `awsext_appintegrations_application`: `Delete` now polls `ListApplicationAssociations` for up to 60s before calling `DeleteApplication`, surfacing a clear error if AWS's async cleanup of Connect-owned `ApplicationAssociation` records hasn't finished, instead of a confusing raw `DeleteApplication` failure.
- `awsext_qconnect_ai_agent_version`: new optional `connect_instance_id` attribute. When set, `Delete` disassociates this version's numbered ARN from any Connect Security Profile still holding it before deleting the version — closes a gap where `awsext_connect_security_profile_association` only manages the base/`$LATEST`/`$SAVED` scopes, never the numbered-version scope, permanently blocking `DeleteSecurityProfile` (`ResourceInUseException`) with no way to clear it short of manual AWS CLI intervention.
- `awsext_qconnect_ai_agent_version`: fixed `Read` double-qualifying the version suffix (`.../<id>:1:1`) when refreshing state — `GetAIAgent` called with an already-versioned `ai_agent_id` returns the ARN already suffixed, unlike `CreateAIAgentVersion`'s response.
- New shared helper `disassociateSecurityProfilesFromEntity` (in `connect_security_profile_association.go`) backing the above fix, using the documented `ListEntitySecurityProfiles` forward lookup rather than parsing AWS error text.

All three fixes were validated against a live AWS account (prod-mock rehearsal environment): confirmed the security profile, integration association, and AppIntegrations application now delete cleanly with zero manual intervention, where they previously required backend AWS Support involvement or manual `disassociate-security-profiles` calls.
