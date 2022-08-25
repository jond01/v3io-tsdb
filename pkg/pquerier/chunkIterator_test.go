// +build integration

/*
Copyright 2018 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/

package pquerier_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/v3io/v3io-tsdb/pkg/config"
	"github.com/v3io/v3io-tsdb/pkg/pquerier"
	"github.com/v3io/v3io-tsdb/pkg/tsdb"
	"github.com/v3io/v3io-tsdb/pkg/tsdb/tsdbtest"
	"github.com/v3io/v3io-tsdb/pkg/utils"
)

const baseTestTime = int64(1547510400000) // 15/01/2019 00:00:00

type testRawChunkIterSuite struct {
	suite.Suite
	v3ioConfig     *config.V3ioConfig
	suiteTimestamp int64
}

func (suite *testRawChunkIterSuite) SetupSuite() {
	v3ioConfig, err := tsdbtest.LoadV3ioConfig()
	suite.Require().NoError(err)

	suite.v3ioConfig = v3ioConfig
	suite.suiteTimestamp = time.Now().Unix()
}

func (suite *testRawChunkIterSuite) SetupTest() {
	suite.v3ioConfig.TablePath = fmt.Sprintf("%s-%v", suite.T().Name(), suite.suiteTimestamp)
	tsdbtest.CreateTestTSDB(suite.T(), suite.v3ioConfig)
}

func (suite *testRawChunkIterSuite) TearDownTest() {
	suite.v3ioConfig.TablePath = fmt.Sprintf("%s-%v", suite.T().Name(), suite.suiteTimestamp)
	if !suite.T().Failed() {
		tsdbtest.DeleteTSDB(suite.T(), suite.v3ioConfig)
	}
}

func (suite *testRawChunkIterSuite) TestRawChunkIteratorWithZeroValue() {
	adapter, err := tsdb.NewV3ioAdapter(suite.v3ioConfig, nil, nil)
	suite.Require().NoError(err)

	labels1 := utils.LabelsFromStringList("os", "linux")
	numberOfEvents := 10
	eventsInterval := 60 * 1000
	ingestData := []tsdbtest.DataPoint{{baseTestTime, 10},
		{baseTestTime + tsdbtest.MinuteInMillis, 0},
		{baseTestTime + 2*tsdbtest.MinuteInMillis, 30},
		{baseTestTime + 3*tsdbtest.MinuteInMillis, 40}}
	testParams := tsdbtest.NewTestParams(suite.T(),
		tsdbtest.TestOption{
			Key: tsdbtest.OptTimeSeries,
			Value: tsdbtest.TimeSeries{tsdbtest.Metric{
				Name:   "cpu",
				Labels: labels1,
				Data:   ingestData},
			}})
	tsdbtest.InsertData(suite.T(), testParams)

	querierV2, err := adapter.QuerierV2()
	suite.Require().NoError(err)

	params, _, _ := pquerier.ParseQuery("select cpu")
	params.From = baseTestTime
	params.To = baseTestTime + int64(numberOfEvents*eventsInterval)

	set, err := querierV2.Select(params)
	suite.Require().NoError(err)

	var seriesCount int
	for set.Next() {
		seriesCount++
		iter := set.At().Iterator().(*pquerier.RawChunkIterator)

		var index int
		for iter.Next() {
			t, v := iter.At()
			prevT, prevV := iter.PeakBack()

			suite.Require().Equal(ingestData[index].Time, t, "current time does not match")

			switch val := ingestData[index].Value.(type) {
			case float64:
				suite.Require().Equal(val, v, "current value does not match")
			case int:
				suite.Require().Equal(float64(val), v, "current value does not match")
			default:
				suite.Require().Equal(val, v, "current value does not match")
			}

			if index > 0 {
				suite.Require().Equal(ingestData[index-1].Time, prevT, "current time does not match")
				switch val := ingestData[index-1].Value.(type) {
				case float64:
					suite.Require().Equal(val, prevV, "current value does not match")
				case int:
					suite.Require().Equal(float64(val), prevV, "current value does not match")
				default:
					suite.Require().Equal(val, prevV, "current value does not match")
				}
			}
			index++
		}
	}

	suite.Require().Equal(1, seriesCount, "series count didn't match expected")
}

func TestRawChunkIterSuite(t *testing.T) {
	suite.Run(t, new(testRawChunkIterSuite))
}
