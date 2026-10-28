package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	undefinedPowerBody      = `{"error":{"code":"undefined_power","message":"power is undefined for these operands"}}`
	resultTooLargeBody      = `{"error":{"code":"result_too_large","message":"result exceeds the supported size"}}`
	negativeSquareRootBody  = `{"error":{"code":"negative_square_root","message":"square root of a negative number is undefined"}}`
	invalidOperandCountBody = `{"error":{"code":"invalid_operand_count","message":"wrong number of operands for operation"}}`
)

// T04 acceptance criteria: power, sqrt and percentage through the unchanged handler and router.
func TestCalculationsAPI_Post_AdvancedOperations(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		want       string
	}{
		// power ["2", "10"] → 1024; power ["2", "-1"] → 0.5; power ["4", "0.5"] → 2
		{
			name: "power 2 10 is 1024", body: `{"operation":"power","operands":["2","10"]}`, wantStatus: http.StatusOK,
			want: `{"operation":"power","operands":["2","10"],"result":"1024"}`,
		},
		{
			name: "power 2 -1 is 0.5", body: `{"operation":"power","operands":["2","-1"]}`, wantStatus: http.StatusOK,
			want: `{"operation":"power","operands":["2","-1"],"result":"0.5"}`,
		},
		{
			name: "power 4 0.5 is 2", body: `{"operation":"power","operands":["4","0.5"]}`, wantStatus: http.StatusOK,
			want: `{"operation":"power","operands":["4","0.5"],"result":"2"}`,
		},

		// power ["0", "-1"], power ["0", "0"], power ["-8", "0.5"] → undefined_power
		{name: "power 0 -1 is undefined", body: `{"operation":"power","operands":["0","-1"]}`, wantStatus: http.StatusUnprocessableEntity, want: undefinedPowerBody},
		{name: "power 0 0 is undefined", body: `{"operation":"power","operands":["0","0"]}`, wantStatus: http.StatusUnprocessableEntity, want: undefinedPowerBody},
		{name: "power -8 0.5 is undefined", body: `{"operation":"power","operands":["-8","0.5"]}`, wantStatus: http.StatusUnprocessableEntity, want: undefinedPowerBody},

		// power ["10", "100000"] → result_too_large
		{name: "power 10 100000 is too large", body: `{"operation":"power","operands":["10","100000"]}`, wantStatus: http.StatusUnprocessableEntity, want: resultTooLargeBody},

		// sqrt ["16"] → 4; sqrt ["2"] → 1.414213562373095; sqrt ["-1"] → negative_square_root; sqrt ["1", "2"] → invalid_operand_count
		{
			name: "sqrt 16 is 4", body: `{"operation":"sqrt","operands":["16"]}`, wantStatus: http.StatusOK,
			want: `{"operation":"sqrt","operands":["16"],"result":"4"}`,
		},
		{
			name: "sqrt 2 rounds to 16 places with trailing zero trimmed", body: `{"operation":"sqrt","operands":["2"]}`, wantStatus: http.StatusOK,
			want: `{"operation":"sqrt","operands":["2"],"result":"1.414213562373095"}`,
		},
		{name: "sqrt -1 is negative_square_root", body: `{"operation":"sqrt","operands":["-1"]}`, wantStatus: http.StatusUnprocessableEntity, want: negativeSquareRootBody},
		{name: "sqrt with 2 operands is invalid_operand_count", body: `{"operation":"sqrt","operands":["1","2"]}`, wantStatus: http.StatusBadRequest, want: invalidOperandCountBody},

		// percentage ["20", "50"] → 10; percentage ["12.5", "80"] → 10
		{
			name: "percentage 20 of 50 is 10", body: `{"operation":"percentage","operands":["20","50"]}`, wantStatus: http.StatusOK,
			want: `{"operation":"percentage","operands":["20","50"],"result":"10"}`,
		},
		{
			name: "percentage 12.5 of 80 is 10", body: `{"operation":"percentage","operands":["12.5","80"]}`, wantStatus: http.StatusOK,
			want: `{"operation":"percentage","operands":["12.5","80"],"result":"10"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newApp(t).calculate(tt.body)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
		})
	}
}

// power ["10", "100000"] → result_too_large (fast, no huge allocation)
func TestCalculationsAPI_Post_HugePowerIsRejectedFast(t *testing.T) {
	a := newApp(t)

	start := time.Now()
	rec := a.calculate(`{"operation":"power","operands":["9.9999999999999999999999999999999999999999999999999","1000"]}`)
	rec2 := a.calculate(`{"operation":"power","operands":["10","100000"]}`)

	assert.Less(t, time.Since(start), 500*time.Millisecond)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, rec2.Code)
}
