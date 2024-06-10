import "./versions.just"

secrets_dir := 'e2e/secrets'
container_provider := env('CONTAINER_PROVIDER', 'podman')
sudo_cmd := env("SUDO", "sudo")
gobin_path := env("GOBIN", "$HOME/.local/bin")
