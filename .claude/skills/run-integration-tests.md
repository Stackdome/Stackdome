# Run Integration Tests

Run integration tests with output saved to a log file for later review.

## Usage

Invoke this skill when you need to run integration tests for the api-server.

## Process

1. **Create timestamped log file**
   - Log files are stored in `test/int/.logs/`
   - Filename format: `int-test-YYYY-MM-DD-HHMMSS.log`

2. **Run the tests**
   - Execute: `go test ./test/int/... -v 2>&1 | tee <log_file>`
   - This runs tests with verbose output and saves to the log file

3. **Report results**
   - Show summary (passed/failed/pending counts)
   - Provide the log file path for later review

## Commands

### Run All Tests (Unit + Integration)

```bash
# Create logs directory if it doesn't exist
mkdir -p test/int/.logs

# Run ALL tests (unit and integration) with timestamped log file
LOG_FILE="test/int/.logs/all-tests-$(date +%Y-%m-%d-%H%M%S).log"
go test ./... -v 2>&1 | tee "$LOG_FILE"
echo "Log saved to: $LOG_FILE"
```

### Run Integration Tests Only

```bash
# Create logs directory if it doesn't exist
mkdir -p test/int/.logs

# Run integration tests with timestamped log file
LOG_FILE="test/int/.logs/int-test-$(date +%Y-%m-%d-%H%M%S).log"
go test ./test/int/... -v 2>&1 | tee "$LOG_FILE"
echo "Log saved to: $LOG_FILE"
```

### Run Unit Tests Only (Short Mode)

```bash
# Run unit tests only (skips integration tests)
LOG_FILE="test/int/.logs/unit-tests-$(date +%Y-%m-%d-%H%M%S).log"
go test ./... -v -short 2>&1 | tee "$LOG_FILE"
echo "Log saved to: $LOG_FILE"
```

## Viewing Logs Later

To view previous test runs:
```bash
# List all log files (most recent first)
ls -lt test/int/.logs/

# View a specific log file
cat test/int/.logs/<filename>.log

# Search for failures in a log
grep -A 10 "FAIL" test/int/.logs/<filename>.log

# View last 100 lines of most recent log
tail -100 test/int/.logs/int-test-*.log | head -100
```

## Running Specific Tests

To run a subset of tests:
```bash
# Run only ObjectStore tests
LOG_FILE="test/int/.logs/objectstore-$(date +%Y-%m-%d-%H%M%S).log"
go test ./test/int/... -v -run="ObjectStore" 2>&1 | tee "$LOG_FILE"

# Run only PostgreSQL addon tests
LOG_FILE="test/int/.logs/postgres-$(date +%Y-%m-%d-%H%M%S).log"
go test ./test/int/... -v -run="PostgreSQL" 2>&1 | tee "$LOG_FILE"
```

## Environment Variables

The tests support these environment variables:
- `TEST_LOG_LEVEL=debug` - Enable debug logging
- `KEEP_CLUSTER=true` - Keep Kind cluster after tests for debugging
- `CLUSTER_AGENT_IMAGE_TAG=latest` - Override cluster agent version

## Notes

- Integration tests require a running PostgreSQL database (configured via .env)
- Tests create a Kind cluster which takes 3-8 minutes for initial setup
- The cluster is automatically cleaned up after tests complete (unless KEEP_CLUSTER=true)
- Unit tests run much faster (seconds) compared to integration tests (minutes)
