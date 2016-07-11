/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/op/go-logging"
	"errors"

	"strings"
	"strconv"
)

var chaincodeLogger = logging.MustGetLogger("blacklist")

// BlacklistChaincode implements the insert into and query of blacklist providing user with a "ally" role.
type BlacklistChaincode struct {
}

var WritesPrefix = "#w"
var ReadsPrefix = "#r"

func (t *BlacklistChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	AllyIds := strings.Split(args[0], ",")
	for _, AllyId := range AllyIds {
		err := stub.PutState(WritesPrefix + AllyId, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, err
		}
		err = stub.PutState(ReadsPrefix + AllyId, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (t *BlacklistChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function == "write" {
		return t.Write(stub, args)
	} else if function == "delete" {
		return t.Delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *BlacklistChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	UserIds := strings.Split(args[0], ",")
	AllyIdBytes, err := stub.GetCallerMetadata()
	NewWrites := 0
	for _, UserId := range UserIds {
		NewWrites++
		err = stub.PutState(UserId, []byte(args[NewWrites]))
		if err != nil {
			return nil, err
		}
	}
	AllyId := string(AllyIdBytes)
	OldWritesBytes, err := stub.GetState(WritesPrefix + AllyId)
	OldWrites, _ := strconv.Atoi(string(OldWritesBytes))
	stub.PutState(AllyId, []byte(strconv.Itoa(OldWrites + NewWrites)))

	return nil, nil
}

func (t *BlacklistChaincode) Delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	UserIds := strings.Split(args[0], ",")
	AllyIdBytes, err := stub.GetCallerMetadata()
	NewWrites := 0
	for _, UserId := range UserIds {
		NewWrites++
		err = stub.DelState(UserId)
		if err != nil {
			return nil, err
		}
	}
	AllyId := string(AllyIdBytes)
	OldWritesBytes, err := stub.GetState(WritesPrefix + AllyId)
	OldWrites, _ := strconv.Atoi(string(OldWritesBytes))
	stub.PutState(AllyId, []byte(strconv.Itoa(OldWrites - NewWrites)))

	return nil, nil
}

func (t *BlacklistChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	UserId := args[0]
	AllyIdBytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Failed getting metadata")
	}

	valAsbytes, err := stub.GetState(UserId)
	if err != nil {
		return nil, err
	}
	AllyId := string(AllyIdBytes)
	OldReadsBytes, err := stub.GetState(ReadsPrefix + AllyId)
	OldReads, _ := strconv.Atoi(string(OldReadsBytes))
	stub.PutState(AllyId, []byte(strconv.Itoa(OldReads + 1)))

	return valAsbytes, nil
}

func main() {
	err := shim.Start(new(BlacklistChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

