package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// GetLocalImportBaseDir returns the root directory under which documents may be
// imported directly from the server's local file system, configured via the
// LOCAL_IMPORT_BASE_DIR environment variable.
//
// When the variable is empty the local-import feature is disabled: both the
// directory-browse and import handlers must reject requests. This is a
// deploy-time-only knob (same rationale as MAX_FILE_SIZE_MB) — the operator is
// responsible for mounting the host directory into the backend container.
//
// The returned path is cleaned but NOT validated for existence here; callers
// pair it with SafePathUnderBase to confine every resolved path under the root.
func GetLocalImportBaseDir() string {
	dir := strings.TrimSpace(os.Getenv("LOCAL_IMPORT_BASE_DIR"))
	if dir == "" {
		return ""
	}
	return filepath.Clean(dir)
}
