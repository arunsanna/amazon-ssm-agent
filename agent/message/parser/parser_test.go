// Copyright 2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// package parser contains utilities for parsing and encoding MDS/SSM messages.
package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aws/amazon-ssm-agent/agent/contracts"
	"github.com/aws/amazon-ssm-agent/agent/log"
	messageContracts "github.com/aws/amazon-ssm-agent/agent/message/contracts"
	"github.com/aws/amazon-ssm-agent/agent/times"
	"github.com/stretchr/testify/assert"
)

var sampleMessageFiles = []string{
	"../testdata/sampleMsg.json",
}

var sampleMessageReplacedParamsFiles = []string{
	"../testdata/sampleMsgReplacedParams.json",
}

var sampleMessageReplyFiles = []string{
	"../testdata/sampleReply.json",
}

var logger = log.NewMockLog()

func TestParseMessageWithParams(t *testing.T) {
	type testCase struct {
		Input  string
		Output messageContracts.SendCommandPayload
	}

	// generate test cases
	var testCases []testCase
	for i, msgFileName := range sampleMessageFiles {
		msgReplacedParamsFileName := sampleMessageReplacedParamsFiles[i]
		testCases = append(testCases, testCase{
			Input:  string(loadFile(t, msgFileName)),
			Output: loadMessageFromFile(t, msgReplacedParamsFileName),
		})
	}

	// run tests
	for _, tst := range testCases {
		// call method
		parsedMsg, err := ParseMessageWithParams(logger, tst.Input)

		// check results
		assert.Nil(t, err)
		assert.Equal(t, tst.Output, parsedMsg)
	}
}

func TestPrepareReplyPayload(t *testing.T) {
	type testCase struct {
		PluginRuntimeStatuses map[string]*contracts.PluginRuntimeStatus
		DateTime              time.Time
		Agent                 contracts.AgentInfo
		Result                messageContracts.SendReplyPayload
	}

	// generate test cases
	var testCases []testCase
	for _, fileName := range sampleMessageReplyFiles {
		// parse a test reply to see if we can regenerate it
		expectedReply := loadMessageReplyFromFile(t, fileName)

		testCases = append(testCases, testCase{
			PluginRuntimeStatuses: expectedReply.RuntimeStatus,
			DateTime:              times.ParseIso8601UTC(expectedReply.AdditionalInfo.DateTime),
			Agent:                 expectedReply.AdditionalInfo.Agent,
			Result:                expectedReply,
		})
	}

	// run test cases
	for _, tst := range testCases {
		// call our method under test
		docResult := PrepareReplyPayload("", tst.PluginRuntimeStatuses, tst.DateTime, tst.Agent)

		// check result
		assert.Equal(t, tst.Result, docResult)
	}
}

func TestPrepareRuntimeStatus(t *testing.T) {
	type testCase struct {
		Input  contracts.PluginResult
		Output contracts.PluginRuntimeStatus
	}

	// generate test cases
	var testCases []testCase
	for _, fileName := range sampleMessageReplyFiles {
		// load the test data
		sampleReply := loadMessageReplyFromFile(t, fileName)

		for _, pluginRuntimeStatus := range sampleReply.RuntimeStatus {
			pluginResult := parsePluginResult(t, *pluginRuntimeStatus)
			testCases = append(testCases, testCase{
				Input:  pluginResult,
				Output: *pluginRuntimeStatus,
			})
		}
	}

	// run test cases
	for _, tst := range testCases {
		// call our method under test
		runtimeStatus := prepareRuntimeStatus(logger, tst.Input)

		// check result
		assert.Equal(t, tst.Output, runtimeStatus)
	}

	// test that there is a runtime status on error
	pluginResult := contracts.PluginResult{Error: fmt.Errorf("Plugin failed with error code 1")}
	runtimeStatus := prepareRuntimeStatus(logger, pluginResult)
	assert.NotNil(t, runtimeStatus.Output)
	return
}

func parsePluginResult(t *testing.T, pluginRuntimeStatus contracts.PluginRuntimeStatus) contracts.PluginResult {
	parsedOutput := pluginRuntimeStatus.Output
	return contracts.PluginResult{
		Output:        parsedOutput,
		Status:        pluginRuntimeStatus.Status,
		StartDateTime: times.ParseIso8601UTC(pluginRuntimeStatus.StartDateTime),
		EndDateTime:   times.ParseIso8601UTC(pluginRuntimeStatus.EndDateTime),
	}
}

func loadFile(t *testing.T, fileName string) (result []byte) {
	result, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func loadMessageFromFile(t *testing.T, fileName string) (message messageContracts.SendCommandPayload) {
	b := loadFile(t, fileName)
	err := json.Unmarshal(b, &message)
	if err != nil {
		t.Fatal(err)
	}
	return message
}

func loadMessageReplyFromFile(t *testing.T, fileName string) (message messageContracts.SendReplyPayload) {
	b := loadFile(t, fileName)
	err := json.Unmarshal(b, &message)
	if err != nil {
		t.Fatal(err)
	}
	return message
}

func Marshalling(t *testing.T) {
	logger := log.Logger()
	defer logger.Flush()
	result := contracts.PluginResult{}

	resultbytes, err := json.Marshal(result)
	if err != nil {
		logger.Info("failed to marshall", err)
	}
	logger.Info("pluginResult=", string(resultbytes))

	result1 := contracts.PluginRuntimeStatus{}

	resultbytes1, err := json.Marshal(result1)
	if err != nil {
		logger.Info("failed to marshall", err)
	}
	logger.Info("PluginRuntimeStatus=", string(resultbytes1))

}
