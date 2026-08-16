package bcn

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FlexInt decodes a JSON number that the API delivers inconsistently:
// buscarjson alternates between a number (10) and a string ("10") for the
// same pagination fields depending on the query (e.g. "Ley 21.600" answers
// strings, "Ley 21461" answers numbers). Every numeric pagination field
// decodes through this type instead of a plain int or string.
type FlexInt int

// UnmarshalJSON accepts a number, a numeric string (trimmed), an integral
// float (10.0 → 10, truncation is deliberate — pagination values never
// carry a fraction), "" and null (both decode to 0, the least surprising
// pagination fallback). Anything else is an explicit error that names the
// offending value: a decode failure must stay visible, never silenced as 0.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	switch {
	case text == "" || text == "null":
		*f = 0
		return nil
	case text[0] == '"':
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("flexint: cannot unquote %s: %w", text, err)
		}
		return f.parse(strings.TrimSpace(s))
	default:
		return f.parse(text)
	}
}

// parse converts the numeric text of one wire field to FlexInt.
func (f *FlexInt) parse(text string) error {
	if text == "" {
		*f = 0
		return nil
	}
	if n, err := strconv.Atoi(text); err == nil {
		*f = FlexInt(n)
		return nil
	}
	if fl, err := strconv.ParseFloat(text, 64); err == nil {
		*f = FlexInt(int(fl))
		return nil
	}
	return fmt.Errorf("flexint: %q is not a number", text)
}

// MarshalJSON emits the value as a JSON number, never a quoted string —
// anything that serializes a Pagination must not revive the API's
// string/number inconsistency.
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(f))), nil
}
