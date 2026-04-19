# Run Integration Tests

Run integration tests using the `make test-integration` target.

## Usage

Invoke this skill when you need to run integration tests for the api-server.

## Process

1. **Run the tests** using `make test-integration`
   - Output is saved to `test/int/last-run.log`
2. **Report results**
   - Show summary (passed/failed/pending counts)
   - Provide the log file path for later review

## Commands

### Run Integration Tests

```bash
make test-integration
```

This runs `go test ./test/int/... -v -ginkgo.v -timeout 30m -count=1` and pipes output to `test/int/last-run.log`.

### Keep Cluster for Debugging

```bash
KEEP_CLUSTER=true make test-integration
```

### With Debug Logging

```bash
TEST_LOG_LEVEL=debug make test-integration
```

### Run Specific Tests (go test directly)

```bash
# Run only PostgreSQL addon tests
go test ./test/int/... -v -ginkgo.v -timeout 30m -count=1 -ginkgo.focus="PostgresAddon"

# Run only e2e lifecycle tests
go test ./test/int/... -v -ginkgo.v -timeout 30m -count=1 -ginkgo.focus="Full Lifecycle|Deletion Cleanup"
```

## Viewing Logs

```bash
# View last run output
cat test/int/last-run.log

# Search for failures
grep -A 10 "FAIL" test/int/last-run.log
```

## Environment Variables

- `TEST_LOG_LEVEL=debug` - Enable debug logging
- `KEEP_CLUSTER=true` - Keep Kind cluster after tests for debugging
- `CLUSTER_AGENT_IMAGE_TAG=latest` - Override cluster agent version

## Notes

- Integration tests require a running PostgreSQL database (configured via .env)
- Tests create a Kind cluster which takes 3-8 minutes for initial setup
- The cluster is automatically cleaned up after tests complete (unless KEEP_CLUSTER=true)
