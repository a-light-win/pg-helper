
import "../lib/test.just"

# Build pg-helper on the host machine
[no-cd]
build-all release_mode='release': (_build_all release_mode)

[no-cd]
build release_mode='release' arch='': (_build release_mode arch)

# Clean the build files
[no-cd]
clean: _clean_build

# Run the unit tests on the host machine
[no-cd]
test: _test_all

# Show the test coverage on the host machine
[no-cd]
coverage: _test_coverage

# Clean the test results
[no-cd]
clean-test: _clean_coverage

# Create new migration
[no-cd]
new-migration name:
  goose -dir internal/db/migrations create {{ name }} sql
