package auth

type AuthConfig struct {
	Jwt  JwtAuthConfig  `embed:"" prefix:"jwt-" help:"JWT auth config"`
	Mtls MtlsAuthConfig `embed:"" prefix:"mtls-" help:"MTLS auth config"`
}
