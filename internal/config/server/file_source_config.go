package server

type FileSourceConfig struct {
	Enabled   bool     `default:"false" negatable:"true" help:"Enable file source"`
	FilePaths []string `validate:"required_if=Enabled true,dive,file|dir" help:"Paths to the source files That declare the databases"`
}
