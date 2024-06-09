
import "../lib/test.just"

# Build pg-helper on the host machine
[no-cd]
build: _build_all

# Clean the build files
[no-cd]
clean: _clean_build
  rm -rf dist/pg-helper

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
