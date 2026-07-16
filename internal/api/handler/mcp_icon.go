// MCP icon presigned upload endpoints. Two mounts share the same handler:
//   - POST /api/v1/mcp/upload/icon           (user auth, mounted in v1 group)
//   - POST /api/v1/admin/mcps/upload/icon    (admin token, mounted at engine root)
//
// Body: {"file_name": "...", "file_size": N, "content_type"?: "image/png"}
// Response: parse.IconUploadResult — presigned PUT URL + persistent download URL
//
// Distinct from the older POST /api/v1/mcps/{id}/icon (multipart, per-record
// stored via WithIconStore). This new flow is presigned URL, unattached, and
// works for BOTH pre-create (frontend uploads then stores URL in the create
// payload) AND edit (same, without needing to hit an /{id}/icon endpoint).

package handler

import (
	"context"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/apierr"
	"github.com/Mininglamp-OSS/octo-marketplace/internal/service/parse"
	"net/http"
)

// McpIconUploadService is the subset of parse.Service this handler needs.
// Kept as an interface so tests can inject fakes without spinning up the
// storage layer.
type McpIconUploadService interface {
	InitMcpIconUpload(ctx context.Context, fileName string, fileSize int64) (*parse.IconUploadResult, error)
}

// McpIcon wires the parse service to a bare-http handler that both the user
// and admin routers can mount.
type McpIcon struct {
	svc McpIconUploadService
}

// NewMcpIcon returns a handler bound to the given service. Nil-safe callers
// should check the return value; a real service is required.
func NewMcpIcon(svc McpIconUploadService) *McpIcon {
	return &McpIcon{svc: svc}
}

type mcpIconInitRequest struct {
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type,omitempty"`
}

// Init handles POST for both mount points. Body validation and error mapping
// mirror the rest of the MCP handler surface (plain JSON, apierr envelope on
// failure) so the octo-admin / octo-web axios clients don't need a special
// case for this endpoint.
func (h *McpIcon) Init(w http.ResponseWriter, r *http.Request) {
	var req mcpIconInitRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.FileName == "" || req.FileSize <= 0 {
		writeError(w, apierr.InvalidRequest("file_name and positive file_size are required"))
		return
	}
	// content_type is informational only right now — the service derives it
	// from the extension. Kept in the request shape so future callers (native
	// clients that already know the type) can be more specific without
	// another API bump.
	result, err := h.svc.InitMcpIconUpload(r.Context(), req.FileName, req.FileSize)
	if err != nil {
		if err == parse.ErrFileTooLarge {
			writeError(w, apierr.InvalidRequest("icon exceeds 2 MiB limit"))
			return
		}
		writeError(w, apierr.InvalidRequest(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, result)
}
