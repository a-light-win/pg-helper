
import "../lib/test.just"

[no-cd,private]
test: _test_all

[no-cd]
coverage: _test_coverage

[no-cd]
clean: _clean_coverage
