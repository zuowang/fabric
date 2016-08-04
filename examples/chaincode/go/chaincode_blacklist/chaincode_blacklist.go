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

	"strconv"
	"bytes"
)

var chaincodeLogger = logging.MustGetLogger("blacklist")

// BlacklistChaincode implements the insert into and query of blacklist providing user with a "Organization" role.
type BlacklistChaincode struct {
}

var WritesPrefix = "#w"
var ReadsPrefix = "#r"

func (t *BlacklistChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting > 0")
	}

	for _, OrganizationId := range args {
		err := stub.PutState(WritesPrefix + OrganizationId, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, err
		}
		err = stub.PutState(ReadsPrefix + OrganizationId, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (t *BlacklistChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	isOk, _ := stub.VerifyAttribute("role", []byte("Organization"))
	if !isOk {
		chaincodeLogger.Errorf("Failed validating caller role")
		return nil, fmt.Errorf("Failed validating caller role.")
	}

	if function == "write" {
		return t.Write(stub, args)
	} else if function == "delete" {
		return t.Delete(stub, args)
	} else if function == "read" {
		return t.Read(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *BlacklistChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	argsLen := len(args)
	if argsLen == 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting > 0")
	}

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, err
	}
	OrganizationId := string(OrganizationIdAsbytes)

	OldWritesAsbytes, err := stub.GetState(WritesPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	OldWrites, err := strconv.Atoi(string(OldWritesAsbytes))
	if err != nil {
		return nil, err
	}
	for i := 1; i < argsLen; i += 2 {
		_, err = stub.GetState(args[i - 1] + OrganizationId)
		if err != nil {
			chaincodeLogger.Errorf("Already exists in blacklist")
			continue
		}

		chaincodeLogger.Infof("PutState [%v]", args[i - 1] + OrganizationId)
		err = stub.PutState(args[i - 1] + OrganizationId, []byte(args[i]))
		if err != nil {
			return nil, err
		}
		OldWrites++
	}
	err = stub.PutState(WritesPrefix + OrganizationId, []byte(strconv.Itoa(OldWrites)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *BlacklistChaincode) Delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting > 0")
	}

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, err
	}
	OrganizationId := string(OrganizationIdAsbytes)
	OldWritesAsbytes, err := stub.GetState(WritesPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	OldWrites, err := strconv.Atoi(string(OldWritesAsbytes))
	if err != nil {
		return nil, err
	}

	for _, UserId := range args {
		err = stub.DelState(UserId + OrganizationId)
		if err != nil {
			return nil, err
		}
		OldWrites--
	}

	err = stub.PutState(WritesPrefix + OrganizationId, []byte(strconv.Itoa(OldWrites)))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *BlacklistChaincode) Read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, err
	}
	OrganizationId := string(OrganizationIdAsbytes)

	UserId := args[0]
	err = stub.PutState(OrganizationId + UserId, []byte(""))
	if err != nil {
		return nil, err
	}

	OldReadsAsbytes, err := stub.GetState(ReadsPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	OldReads, err := strconv.Atoi(string(OldReadsAsbytes))
	if err != nil {
		return nil, err
	}
	err = stub.PutState(ReadsPrefix + OrganizationId, []byte(strconv.Itoa(OldReads + 1)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *BlacklistChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	isOk, _ := stub.VerifyAttribute("role", []byte("Organization"))
	if !isOk {
		chaincodeLogger.Errorf("Failed validating caller role")
		return nil, fmt.Errorf("Failed validating caller role.")
	}

	if function == "fetch" {
		return t.Fetch(stub, args)
	} else if function == "account" {
		return t.Account(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *BlacklistChaincode) Fetch(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, err
	}
	OrganizationId := string(OrganizationIdAsbytes)

	UserId := args[0]
	err = stub.GetState(OrganizationId + UserId)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	iter, err := stub.RangeQueryState(UserId + "$", UserId + "~")
	if err != nil {
		return nil, fmt.Errorf("Error fetching blacklist: [%v]", err)
	}
	defer iter.Close()

	for iter.HasNext() {
		_, valBytes, err := iter.Next()
		if err != nil {
			return nil, err
		}
		buffer.Write(valBytes)
		buffer.WriteString("|")
	}

	return buffer.Bytes(), nil
}

func (t *BlacklistChaincode) Account(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, err
	}
	OrganizationId := string(OrganizationIdAsbytes)

	var buffer bytes.Buffer
	writesBytes, err := stub.GetState(WritesPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	buffer.Write(writesBytes)
	buffer.WriteString("|")
	readsBytes, err := stub.GetState(ReadsPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	buffer.Write(readsBytes)

	return buffer.Bytes(), nil
}

func main() {
	err := shim.Start(new(BlacklistChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

