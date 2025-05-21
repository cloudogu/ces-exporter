package configuration

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertTimeToCron(t *testing.T) {
	tests := []struct {
		time           string
		expectedResult string
	}{
		{
			time:           "0:0",
			expectedResult: "0 0 * * *",
		},
		{
			time:           "00:00",
			expectedResult: "00 00 * * *",
		},
		{
			time:           "01:01",
			expectedResult: "01 01 * * *",
		},
		{
			time:           "01:02",
			expectedResult: "02 01 * * *",
		},
		{
			time:           "23:59",
			expectedResult: "59 23 * * *",
		},
		{
			time:           "12:00",
			expectedResult: "00 12 * * *",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("convert %s correctly", test.time), func(t *testing.T) {
			assert.Equal(t, test.expectedResult, convertTimeToCron(test.time))
		})
	}
}
