package bcn

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

// FlexIntSuite validates the lenient numeric decoding against every wire
// shape buscarjson has shown: numbers, strings, and their mix.
type FlexIntSuite struct {
	suite.Suite
}

func TestFlexIntSuite(t *testing.T) {
	suite.Run(t, new(FlexIntSuite))
}

func (s *FlexIntSuite) TestUnmarshalWires() {
	cases := []struct {
		name    string
		wire    string
		want    FlexInt
		wantErr bool
	}{
		{name: "number", wire: `10`, want: 10},
		{name: "negative number", wire: `-1`, want: -1},
		{name: "string", wire: `"10"`, want: 10},
		{name: "string with spaces", wire: `" 10 "`, want: 10},
		{name: "float number", wire: `10.0`, want: 10},
		{name: "float truncated", wire: `10.9`, want: 10}, // pagination never carries a fraction
		{name: "float string", wire: `"10.0"`, want: 10},
		{name: "zero number", wire: `0`, want: 0},
		{name: "empty string", wire: `""`, want: 0},
		{name: "null", wire: `null`, want: 0},
		{name: "non-numeric string", wire: `"abc"`, wantErr: true},
		{name: "bool", wire: `true`, wantErr: true},
		{name: "array", wire: `[]`, wantErr: true},
		{name: "object", wire: `{}`, wantErr: true},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			var got FlexInt
			err := json.Unmarshal([]byte(tc.wire), &got)
			if tc.wantErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *FlexIntSuite) TestErrorNamesOffendingValue() {
	var got FlexInt
	err := json.Unmarshal([]byte(`"abc"`), &got)
	s.Require().Error(err)
	s.Contains(err.Error(), `"abc"`, "error must name the offending value")
}

func (s *FlexIntSuite) TestMarshalEmitsNumber() {
	data, err := json.Marshal(FlexInt(140))
	s.Require().NoError(err)
	s.Equal("140", string(data), "marshaled value is a number, never a quoted string")
}

func (s *FlexIntSuite) TestRoundTripStable() {
	// The same value arriving as a string or as a number marshals to the
	// same bytes: consumers never see the API's inconsistency again.
	var fromString, fromNumber FlexInt
	s.Require().NoError(json.Unmarshal([]byte(`"10"`), &fromString))
	s.Require().NoError(json.Unmarshal([]byte(`10`), &fromNumber))
	s.Equal(fromNumber, fromString)

	strBytes, err := json.Marshal(fromString)
	s.Require().NoError(err)
	numBytes, err := json.Marshal(fromNumber)
	s.Require().NoError(err)
	s.Equal(string(numBytes), string(strBytes))
}
