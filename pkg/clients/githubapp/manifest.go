package githubapp

// GitHub App permission scopes and access levels requested by the manifest.
const (
	PermContents  = "contents"
	PermMetadata  = "metadata"
	PermLevelRead = "read"
)

// AppManifest is the GitHub App creation manifest the browser POSTs to
// github.com/settings/apps/new. GitHub has no SDK type for this input (go-github
// only converts the resulting code), so it is defined here alongside the other
// GitHub App wire types. The hub-specific URLs are filled in by the caller.
type AppManifest struct {
	Name               string            `json:"name"`
	URL                string            `json:"url"`
	RedirectURL        string            `json:"redirect_url"`
	SetupURL           string            `json:"setup_url"`
	Public             bool              `json:"public"`
	HookAttributes     AppHookAttributes `json:"hook_attributes"`
	DefaultPermissions map[string]string `json:"default_permissions"`
	DefaultEvents      []string          `json:"default_events"`
}

// AppHookAttributes configures where GitHub delivers the app's webhooks.
type AppHookAttributes struct {
	URL string `json:"url"`
}
