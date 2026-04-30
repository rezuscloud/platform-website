package apps

import (
	"encoding/json"

	model "github.com/rezuscloud/platform-website/internal/platform"
)

func jsonHistory(state model.SessionState) string {
	data, err := json.Marshal(state.Terminal.History)
	if err != nil {
		return "[]"
	}
	return string(data)
}
