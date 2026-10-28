package tests

import (
	"bufio"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T10: the access log carries the API error code so failures can be filtered by contract code in Loki.
func TestAccessLog_RecordsTheAPIErrorCode(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantStatus    int
		wantErrorCode string
	}{
		{name: "division_by_zero is logged with its code", body: `{"operation":"divide","operands":["10","0"]}`, wantStatus: http.StatusUnprocessableEntity, wantErrorCode: "division_by_zero"},
		{name: "invalid_request is logged with its code", body: `{`, wantStatus: http.StatusBadRequest, wantErrorCode: "invalid_request"},
		{name: "a success has no error_code", body: `{"operation":"add","operands":["2","3"]}`, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newApp(t)

			a.calculate(tt.body)

			entry := accessLogEntry(t, a)
			assert.InDelta(t, tt.wantStatus, entry["status"], 0)
			if tt.wantErrorCode == "" {
				assert.NotContains(t, entry, "error_code")
				return
			}
			assert.Equal(t, tt.wantErrorCode, entry["error_code"])
		})
	}
}

func accessLogEntry(t *testing.T, a app) map[string]any {
	t.Helper()
	scanner := bufio.NewScanner(a.logs)
	for scanner.Scan() {
		var entry map[string]any
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &entry))
		if entry["msg"] == "request" {
			return entry
		}
	}
	require.FailNow(t, "no access log line", a.logs.String())
	return nil
}
