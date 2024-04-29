package auth

type JwtAuthConfig struct {
	Enabled  bool   `default:"false"`
	Audience string `help:"The client name registered in the JWT provider"`

	JwtVerifyKey `embed:"" prefix:"verify-key-" validate:"required_if=Enabled true"`
}
