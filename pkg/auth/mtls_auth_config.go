package auth

type MtlsAuthConfig struct {
	Enabled bool `negatable:"true" help:"Enable mutual tls auth"`
	// TrustedCA  string `name:"trusted-ca" validate:"required_if=Enabled true,omitempty,file" help:"Path to the server trusted ca certs"`
	BaseDomain string `help:"The base domain of client cert"`

	ScopeEnabled    bool `negatable:"true" help:"Should we load scopes from client cert"`
	ResourceEnabled bool `negatable:"true" help:"Should we load resources from client cert"`
}
