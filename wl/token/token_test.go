package token

import (
	"strings"
	"testing"
)

func TestTokenToName(t *testing.T) {
	// make sure all our tokens are accounted for in the naming list
	validateRange(t, -1, literal_beg)
	validateRange(t, literal_beg, literal_end)
	validateRange(t, operator_beg, operator_end)
	validateRange(t, keyword_beg, keyword_end)
}

func validateRange(t *testing.T, startAfter, endBefore Token) {
	for i := startAfter + 1; i < endBefore; i++ {
		val := i.String()
		if val == "" || strings.HasPrefix(val, "token(") {
			t.Errorf("unknown token: %v", i)
		}
	}
}
