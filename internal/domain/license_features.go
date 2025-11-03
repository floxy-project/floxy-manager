package domain

// LicenseFeature represents a feature that can be enabled or disabled based on license type.
type LicenseFeature string

const (
	// FeatureSSO represents SSO/SAML authentication feature.
	FeatureSSO LicenseFeature = "sso"

	// FeatureLDAP represents LDAP authentication and synchronization feature.
	FeatureLDAP LicenseFeature = "ldap"

	// (Slack, Pachca, Mattermost, Webhooks).
	FeatureCorpNotifChannels LicenseFeature = "corp_notif_channels"
)

// GetAvailableFeatures returns the list of features available for the given license type.
func GetAvailableFeatures(licenseType LicenseType) []LicenseFeature {
	switch licenseType {
	case Trial, TrialSelfSigned, Commercial:
		// All features are available for trial, trial self-signed, and commercial licenses
		return []LicenseFeature{
			FeatureSSO,
			FeatureLDAP,
			FeatureCorpNotifChannels,
		}
	default:
		// Default to no features for unknown license types
		return []LicenseFeature{}
	}
}

// IsFeatureAvailable checks if a specific feature is available for the given license type.
func IsFeatureAvailable(licenseType LicenseType, feature LicenseFeature) bool {
	availableFeatures := GetAvailableFeatures(licenseType)
	for _, f := range availableFeatures {
		if f == feature {
			return true
		}
	}

	return false
}
