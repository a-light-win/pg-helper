secrets_dir := 'e2e/secrets'
container_provider := env('CONTAINER_PROVIDER', 'podman')
sudo_cmd := env("SUDO", "sudo")
gobin_path := env("GOBIN", "~/.local/bin")

protoc_gen_go_version := "v1.33.0"
protoc_gen_go_grpc_version := "v1.3.0"
sqlc_version := "v1.26.0"
goose_version := "v3.20.0"
