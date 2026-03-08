package dashboard

import (
	"encoding/base64"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// apiRequest is the JSON representation of a store.Request sent to the
// dashboard. It embeds all exported fields plus base64-encoded request/
// response bodies (excluded from store serialisation with json:"-").
type apiRequest struct {
	*store.Request
	RequestBody  string `json:"request_body,omitempty"`
	ResponseBody string `json:"response_body,omitempty"`
}

func toDTO(r *store.Request) apiRequest {
	dto := apiRequest{Request: r}
	if len(r.RequestBody) > 0 {
		dto.RequestBody = base64.StdEncoding.EncodeToString(r.RequestBody)
	}
	if len(r.ResponseBody) > 0 {
		dto.ResponseBody = base64.StdEncoding.EncodeToString(r.ResponseBody)
	}
	return dto
}
