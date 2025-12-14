# K8SPACKET REPOSITORY RESTRUCTURING - MIGRATION PHASES

## Overview
Complete repository restructuring according to Go best practices:
- Move entry points to `cmd/`
- Create public API in `pkg/`
- Move infrastructure to `internal/`

---

## PHASE 1: Move Entry Points to cmd/ (root → cmd/k8spacket/)

**Objective:** Consolidate root `k8spacket.go` and `k8spacket_test.go` into `cmd/k8spacket/`.

**Files to move:**
- `k8spacket.go` → `cmd/k8spacket/k8spacket.go`
- `k8spacket_test.go` → `cmd/k8spacket/k8spacket_test.go`

**Actions:**
1. Copy root files to cmd/k8spacket/
2. Remove root files
3. Verify build and tests pass

**Verification:**
- `go build ./cmd/k8spacket/` ✓
- `go test ./cmd/k8spacket/...` ✓
- Root files removed

---

## PHASE 2: Create Public API (pkg/events, pkg/api)

**Objective:** Extract public models and interfaces to `pkg/` for external consumption.

**Directories to create:**
- `pkg/events/`
- `pkg/api/`

**Files to create:**

### pkg/events/model.go
- Copy from `modules/model.go`
- Change package to `package events`
- Export: `TCPEvent`, `TLSEvent`, `Address`, `EventSource`, `Connection`

### pkg/api/listener.go
- Copy from `modules/ilistener.go`
- Change package to `package api`
- Import: `github.com/k8spacket/k8spacket/pkg/events`
- Export: `IListener[T]`

### pkg/api/broker.go (NEW)
- Create minimal interface
- Define `IBroker` interface
- Export public broker contract

**Verification:**
- `go test ./pkg/...` ✓
- `go build ./pkg/...` ✓

---

## PHASE 3: Move eBPF & Broker to internal/

**Objective:** Move infrastructure (eBPF, broker) to `internal/` with import updates.

**Directories to create:**
- `internal/ebpf/inet/`
- `internal/ebpf/socketfilter/`
- `internal/ebpf/tc/`
- `internal/ebpf/tools/`
- `internal/broker/`

**Files to move (git mv):**
- `ebpf/iloader.go` → `internal/ebpf/iloader.go`
- `ebpf/loader.go` → `internal/ebpf/loader.go`
- `ebpf/loader_test.go` → `internal/ebpf/loader_test.go`
- `ebpf/inet/*` → `internal/ebpf/inet/`
- `ebpf/socketfilter/*` → `internal/ebpf/socketfilter/`
- `ebpf/tc/*` → `internal/ebpf/tc/`
- `ebpf/tools/*` → `internal/ebpf/tools/`
- `broker/*` → `internal/broker/`

**Cleanup:**
- Remove empty `ebpf/` and `broker/` directories

**Import changes in cmd/k8spacket/k8spacket.go:**
```
github.com/k8spacket/k8spacket/broker                 → github.com/k8spacket/k8spacket/internal/broker
github.com/k8spacket/k8spacket/ebpf                   → github.com/k8spacket/k8spacket/internal/ebpf
github.com/k8spacket/k8spacket/ebpf/inet              → github.com/k8spacket/k8spacket/internal/ebpf/inet
github.com/k8spacket/k8spacket/ebpf/socketfilter      → github.com/k8spacket/k8spacket/internal/ebpf/socketfilter
github.com/k8spacket/k8spacket/ebpf/tc                → github.com/k8spacket/k8spacket/internal/ebpf/tc
```

**Import changes in internal/ebpf/* and internal/broker/*:**
```
github.com/k8spacket/k8spacket/modules     → github.com/k8spacket/k8spacket/pkg/events
github.com/k8spacket/k8spacket/broker      → github.com/k8spacket/k8spacket/internal/broker
```

**Verification:**
- `go mod tidy` ✓
- `go build ./cmd/k8spacket/` ✓
- `go test ./internal/...` ✓
- `go test ./cmd/k8spacket/` ✓

---

## PHASE 4: Move external/ to internal/infra/

**Objective:** Move external service adapters to `internal/infra/`.

**Directories to create:**
- `internal/infra/db/`
- `internal/infra/http/`
- `internal/infra/k8s/`
- `internal/infra/network/`
- `internal/infra/handlerio/`

**Files to move (git mv):**
- `external/db/*` → `internal/infra/db/`
- `external/http/*` → `internal/infra/http/`
- `external/k8s/*` → `internal/infra/k8s/`
- `external/network/*` → `internal/infra/network/`
- `external/handlerio/*` → `internal/infra/handlerio/`

**Cleanup:**
- Remove empty `external/` directory

**Import changes in modules/* (all .go files):**
```
github.com/k8spacket/k8spacket/external/db          → github.com/k8spacket/k8spacket/internal/infra/db
github.com/k8spacket/k8spacket/external/http        → github.com/k8spacket/k8spacket/internal/infra/http
github.com/k8spacket/k8spacket/external/k8s         → github.com/k8spacket/k8spacket/internal/infra/k8s
github.com/k8spacket/k8spacket/external/network     → github.com/k8spacket/k8spacket/internal/infra/network
github.com/k8spacket/k8spacket/external/handlerio   → github.com/k8spacket/k8spacket/internal/infra/handlerio
```

**Verification:**
- `go mod tidy` ✓
- `go test ./modules/...` ✓
- `go test ./internal/infra/...` ✓

---

## PHASE 5: Move modules to internal/plugins

**Objective:** Move application modules (nodegraph, tls-parser) to `internal/plugins/`.

**Directories to create:**
- `internal/plugins/nodegraph/`
- `internal/plugins/tls-parser/`
- (all subdirectories: model/, stats/, repository/, prometheus/, certificate/, dict/)

**Files to move (git mv):**
- `modules/nodegraph/*` → `internal/plugins/nodegraph/`
- `modules/tls-parser/*` → `internal/plugins/tls-parser/`

**Cleanup:**
- Remove empty `modules/` directory

**Import changes in cmd/k8spacket/k8spacket.go:**
```
github.com/k8spacket/k8spacket/modules/nodegraph    → github.com/k8spacket/k8spacket/internal/plugins/nodegraph
github.com/k8spacket/k8spacket/modules/tls-parser   → github.com/k8spacket/k8spacket/internal/plugins/tls-parser
github.com/k8spacket/k8spacket/modules              → github.com/k8spacket/k8spacket/pkg/events
```

**Import changes in internal/plugins/**/*.go:**
```
github.com/k8spacket/k8spacket/modules/nodegraph    → github.com/k8spacket/k8spacket/internal/plugins/nodegraph
github.com/k8spacket/k8spacket/modules/tls-parser   → github.com/k8spacket/k8spacket/internal/plugins/tls-parser
github.com/k8spacket/k8spacket/modules              → github.com/k8spacket/k8spacket/pkg/events
github.com/k8spacket/k8spacket/external/*           → github.com/k8spacket/k8spacket/internal/infra/*
```

**Verification:**
- `go mod tidy` ✓
- `go test ./internal/plugins/...` ✓
- `go build ./cmd/k8spacket/` ✓

---

## PHASE 6: Final Verification & Cleanup

**Objective:** Comprehensive verification, testing, formatting.

**Step 6.1: Verify old imports are gone**
```bash
grep -r "github.com/k8spacket/k8spacket/external/" --include="*.go" . | grep -v ".git/"
# Expected: ZERO matches

grep -r "github.com/k8spacket/k8spacket/ebpf" --include="*.go" . | grep -v ".git/" | grep -v "internal/ebpf"
# Expected: ZERO matches outside internal/

grep -r "github.com/k8spacket/k8spacket/broker" --include="*.go" . | grep -v ".git/" | grep -v "internal/broker"
# Expected: ZERO matches outside internal/

grep -r "github.com/k8spacket/k8spacket/modules" --include="*.go" . | grep -v ".git/" | grep -v "internal/plugins" | grep -v "pkg/"
# Expected: ZERO matches outside plugins/ or pkg/
```

**Step 6.2: Run all tests**
```bash
go test ./...
# Expected: ALL tests PASS

go test -count=5 ./...
# Multiple runs to catch race conditions
```

**Step 6.3: Format code**
```bash
gofmt -s -w .
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

**Step 6.4: Tidy dependencies**
```bash
go mod tidy
go mod verify
```

**Step 6.5: Build binary**
```bash
cd cmd/k8spacket/ && go build -o k8spacket . && cd -
# Expected: Binary builds successfully
```

**Step 6.6: Verify directory structure**
```bash
tree -d -L 3 --dirsfirst
# Expected:
# cmd/k8spacket/
# internal/
#   broker/
#   ebpf/
#     inet/
#     socketfilter/
#     tc/
#     tools/
#   infra/
#     db/
#     handlerio/
#     http/
#     k8s/
#     network/
#   plugins/
#     nodegraph/
#     tls-parser/
# pkg/
#   api/
#   events/
# tests/
# dashboards/
# docs/
```

**Step 6.7: Check no orphaned directories**
```bash
ls -la ebpf/ 2>/dev/null && echo "❌ ERROR: ebpf/ still exists" || echo "✓ ebpf/ removed"
ls -la broker/ 2>/dev/null && echo "❌ ERROR: broker/ still exists" || echo "✓ broker/ removed"
ls -la external/ 2>/dev/null && echo "❌ ERROR: external/ still exists" || echo "✓ external/ removed"
ls -la modules/ 2>/dev/null && echo "❌ ERROR: modules/ still exists" || echo "✓ modules/ removed"
```

**Step 6.8: Verify go.mod unchanged**
```bash
head -1 go.mod
# Expected: "module github.com/k8spacket/k8spacket"
```

**Verification Checklist:**
- ✓ ZERO old import paths remain
- ✓ ALL tests pass
- ✓ NO compilation errors
- ✓ Directory structure clean
- ✓ go.mod tidy and verified
- ✓ Binary builds successfully
- ✓ Public API in pkg/
- ✓ Internal code in internal/

---

## Import Mapping Summary

| Old Import Path | New Import Path |
|---|---|
| `github.com/k8spacket/k8spacket/modules` | `github.com/k8spacket/k8spacket/pkg/events` |
| `github.com/k8spacket/k8spacket/broker` | `github.com/k8spacket/k8spacket/internal/broker` |
| `github.com/k8spacket/k8spacket/ebpf` | `github.com/k8spacket/k8spacket/internal/ebpf` |
| `github.com/k8spacket/k8spacket/ebpf/inet` | `github.com/k8spacket/k8spacket/internal/ebpf/inet` |
| `github.com/k8spacket/k8spacket/ebpf/socketfilter` | `github.com/k8spacket/k8spacket/internal/ebpf/socketfilter` |
| `github.com/k8spacket/k8spacket/ebpf/tc` | `github.com/k8spacket/k8spacket/internal/ebpf/tc` |
| `github.com/k8spacket/k8spacket/ebpf/tools` | `github.com/k8spacket/k8spacket/internal/ebpf/tools` |
| `github.com/k8spacket/k8spacket/external/db` | `github.com/k8spacket/k8spacket/internal/infra/db` |
| `github.com/k8spacket/k8spacket/external/http` | `github.com/k8spacket/k8spacket/internal/infra/http` |
| `github.com/k8spacket/k8spacket/external/k8s` | `github.com/k8spacket/k8spacket/internal/infra/k8s` |
| `github.com/k8spacket/k8spacket/external/network` | `github.com/k8spacket/k8spacket/internal/infra/network` |
| `github.com/k8spacket/k8spacket/external/handlerio` | `github.com/k8spacket/k8spacket/internal/infra/handlerio` |
| `github.com/k8spacket/k8spacket/modules/nodegraph` | `github.com/k8spacket/k8spacket/internal/plugins/nodegraph` |
| `github.com/k8spacket/k8spacket/modules/tls-parser` | `github.com/k8spacket/k8spacket/internal/plugins/tls-parser` |

---

## Estimated Execution Time
- Phase 1: 3-5 min
- Phase 2: 5-10 min
- Phase 3: 10-15 min
- Phase 4: 10-15 min
- Phase 5: 15-20 min
- Phase 6: 5-10 min
- **Total: ~50-75 min**

---

## Ready for Execution
All phases are documented and ready to execute. Each phase includes:
- Specific files/directories to move
- Exact import changes required
- Verification steps
- Expected outcomes

Start with Phase 1 when ready.