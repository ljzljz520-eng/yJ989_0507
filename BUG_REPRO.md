# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	activityregistration/cmd/activity	[no test files]
?   	activityregistration/internal/clock	[no test files]
ok  	activityregistration/internal/domain	0.001s
ok  	activityregistration/internal/exporter	0.002s
ok  	activityregistration/internal/httpapi	0.017s
ok  	activityregistration/internal/importer	0.001s
ok  	activityregistration/internal/integration	0.033s
--- FAIL: TestBusiness014Regression (0.07s)
    regression_test.go:47: page two lost continuity at offset 0: got r-031 want r-030
FAIL
FAIL	activityregistration/internal/service	0.080s
ok  	activityregistration/internal/store	0.020s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/activity): exit `0`
- Frontend build (web): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/activity): exit `0`
- Frontend build (web): exit `0`
