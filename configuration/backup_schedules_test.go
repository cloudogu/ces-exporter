package configuration

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGetBackupSchedulesFunc(t *testing.T) {
	t.Run("will work for valid time", func(t *testing.T) {
		provider := newBackupScheduleProvider(func(dogu string, ignoreKeys []string) (core.DoguConfig, error) {
			return core.DoguConfig{
				NormalConfig: []core.KeyValue{
					{
						Key:   "/time",
						Value: "13:30",
					},
				},
			}, nil
		})

		schedules, err := provider.getBackupSchedules()
		require.NoError(t, err)

		require.Equal(t, 1, len(schedules))
		assert.Equal(t, "30 13 * * *", schedules[0].Schedule)
		assert.Equal(t, scheduledBackupName, schedules[0].Name)
	})
	t.Run("will return empty array on invalid time", func(t *testing.T) {
		provider := newBackupScheduleProvider(func(dogu string, ignoreKeys []string) (core.DoguConfig, error) {
			return core.DoguConfig{
				NormalConfig: []core.KeyValue{
					{
						Key:   "/time",
						Value: "1330",
					},
				},
			}, nil
		})

		schedules, err := provider.getBackupSchedules()
		require.NoError(t, err)

		require.Equal(t, 0, len(schedules))
	})
	t.Run("will return error if error occur on get config", func(t *testing.T) {
		provider := newBackupScheduleProvider(func(dogu string, ignoreKeys []string) (core.DoguConfig, error) {
			return core.DoguConfig{}, fmt.Errorf("testerror")
		})

		schedules, err := provider.getBackupSchedules()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "testerror")
		assert.Equal(t, 0, len(schedules))
	})
}
