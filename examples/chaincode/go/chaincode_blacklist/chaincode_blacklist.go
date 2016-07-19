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
	"time"
)

var chaincodeLogger = logging.MustGetLogger("blacklist")

// BlacklistChaincode implements the insert into and query of blacklist providing user with a "Organization" role.
type BlacklistChaincode struct {
}

var WritesPrefix = "#w"
var ReadsPrefix = "#r"
var SharesPrefix = "#s"
var CreditsPrefix = "#c"
var LeaseStartPrefix = "#ls"
var LeaseEndPrefix = "#ls"
var SharesMiddle = "#"

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
		err = stub.PutState(SharesPrefix + OrganizationId, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, err
		}
		err = stub.PutState(CreditsPrefix + OrganizationId, []byte(strconv.Itoa(950)))
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (t *BlacklistChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	callerRole, err := stub.ReadCertAttribute("role")
	if err != nil {
		chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
		return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	}

	caller := string(callerRole)
	if (caller != "Organization") {
		chaincodeLogger.Errorf("Failed validating caller role")
		return nil, fmt.Errorf("Failed validating caller role.")
	}
	if function == "write" {
		return t.Write(stub, args)
	} else if function == "delete" {
		return t.Delete(stub, args)
	} else if function == "read" {
		return t.Read(stub, args)
	} else if function == "lease" {
		return t.Lease(stub, args)
	} else if function == "lease2" {
		return t.Lease2(stub, args)
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
	//OrganizationId := ""
	//OldWrites := 0
	//var err error
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

		chaincodeLogger.Infof("PutState [%v]", args[i - 1] + SharesMiddle + OrganizationId)
		err = stub.PutState(args[i - 1] + SharesMiddle + OrganizationId, []byte(strconv.Itoa(0)))
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
	OldSharesAsbytes, err := stub.GetState(SharesPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	OldShares, err := strconv.Atoi(string(OldSharesAsbytes))
	if err != nil {
		return nil, err
	}
	for _, UserId := range args {
		err = stub.DelState(UserId + OrganizationId)
		if err != nil {
			return nil, err
		}
		OldWrites--
		sharesBytes, err := stub.GetState(UserId + SharesMiddle + OrganizationId)
		if err != nil {
			return nil, err
		}
		shares, err := strconv.Atoi(string(sharesBytes))
		OldShares -= shares
		err = stub.DelState(UserId + SharesMiddle + OrganizationId)
		if err != nil {
			return nil, err
		}
	}

	err = stub.PutState(WritesPrefix + OrganizationId, []byte(strconv.Itoa(OldWrites)))
	if err != nil {
		return nil, err
	}
	err = stub.PutState(SharesPrefix + OrganizationId, []byte(strconv.Itoa(OldShares)))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *BlacklistChaincode) Read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	callerRole, err := stub.ReadCertAttribute("role")
	if err != nil {
		chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
		return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	}

	caller := string(callerRole)
	if (caller != "Organization") {
		chaincodeLogger.Errorf("Failed validating caller role")
		return nil, fmt.Errorf("Failed validating caller role.")
	}

	UserId := args[0]
	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Failed getting metadata")
	}
	OrganizationId := string(OrganizationIdAsbytes)

	iter, err := stub.RangeQueryState(UserId+"$", UserId+"~")
	if err != nil {
		return nil, fmt.Errorf("Error fetching blacklist: [%v]", err)
	}
	defer iter.Close()

	var buffer bytes.Buffer
	if !iter.HasNext() {
		return nil, fmt.Errorf("Fetch nil for user: [%v]", UserId)
	}

	OldReadsAsbytes, err := stub.GetState(ReadsPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	OldReads, err := strconv.Atoi(string(OldReadsAsbytes))
	if err != nil {
		return nil, err
	}

	for iter.HasNext() {
		keyStr, valBytes, err := iter.Next()
		if err != nil {
			return nil, err
		}
		buffer.Write(valBytes)
		buffer.WriteString("|")

		val := keyStr[18:]
		// Read counter plus one if read other's blacklist
		if (val != OrganizationId) {
			OldReads++

			OldSharesAsbytes, err := stub.GetState(SharesPrefix + val)
			if err != nil {
				return nil, err
			}
			OldShares, err := strconv.Atoi(string(OldSharesAsbytes))
			if err != nil {
				return nil, err
			}
			err = stub.PutState(SharesPrefix + val, []byte(strconv.Itoa(OldShares + 1)))
			if err != nil {
				return nil, err
			}

			OldSharesAsbytes, err = stub.GetState(UserId + SharesMiddle + val)
			if err != nil {
				return nil, err
			}
			OldShares, err = strconv.Atoi(string(OldSharesAsbytes))
			if err != nil {
				return nil, err
			}
			err = stub.PutState(UserId + SharesMiddle + val, []byte(strconv.Itoa(OldShares + 1)))
			if err != nil {
				return nil, err
			}
		}
	}

	err = stub.PutState(OrganizationId, []byte(strconv.Itoa(OldReads)))
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (t *BlacklistChaincode) Lease(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	callerRole, err := stub.ReadCertAttribute("role")
	if err != nil {
		chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
		return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	}

	caller := string(callerRole)
	if (caller != "Organization") {
		chaincodeLogger.Errorf("Failed validating caller role")
		return nil, fmt.Errorf("Failed validating caller role.")
	}

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Failed getting metadata")
	}
	OrganizationId := string(OrganizationIdAsbytes)

	err = stub.PutState(LeaseStartPrefix + OrganizationId, []byte(time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *BlacklistChaincode) Lease2(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	callerRole, err := stub.ReadCertAttribute("role")
	if err != nil {
		chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
		return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	}

	caller := string(callerRole)
	if (caller != "Organization") {
		chaincodeLogger.Errorf("Failed validating caller role")
		return nil, fmt.Errorf("Failed validating caller role.")
	}

	secondStr := args[0]
	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Failed getting metadata")
	}
	OrganizationId := string(OrganizationIdAsbytes)

	seconds, err := strconv.Atoi(secondStr)
	if err != nil {
		return nil, err
	}

	lease := time.Duration(seconds) * time.Second
	leaseEndtime := time.Now().Add(lease)
	err = stub.PutState(LeaseEndPrefix + OrganizationId, []byte(leaseEndtime.Format("2006-01-02 15:04:05")))
	if err != nil {
		return nil, err
	}

	// Initiate Timer for the duration of the lease
	go func(stub *shim.ChaincodeStub, OrganizationId string, sleeptime time.Duration) ([]byte, error) {
		fmt.Println("Lease2: Sleeping for ", sleeptime)
		//time.Sleep(sleeptime)
		err := stub.DelState(LeaseEndPrefix + OrganizationId)
		if err != nil {
			fmt.Println("Del State Fail===========================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================%v", err)
			return nil, err
		}
		return nil, nil
	}(stub, OrganizationId, lease)
	return nil, nil
}

func (t *BlacklistChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	callerRole, err := stub.ReadCertAttribute("role")
	if err != nil {
		chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
		return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	}

	caller := string(callerRole)
	if (caller != "Organization") {
		chaincodeLogger.Errorf("Failed validating caller role")
		return nil, fmt.Errorf("Failed validating caller role.")
	}

	if function == "fetch" {
		return t.Fetch(stub, args)
	} else if function == "account" {
		return t.Account(stub, args)
	} else if function == "fetch2" {
		return t.Fetch2(stub, args)
	} else if function == "fetch3" {
		return t.Fetch3(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *BlacklistChaincode) Fetch(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	UserId := args[0]

	var buffer bytes.Buffer
	//valBytes, err := stub.GetState(UserId + "idc")
	//if err != nil {
	//	return nil, err
	//}
	//buffer.Write(valBytes)
	//buffer.WriteString("|")
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

	//OrganizationId := ""
	//chaincodeLogger.Infof("GetState [%v]", UserId + OrganizationId)
	//resultBytes, err := stub.GetState(UserId + OrganizationId)
	//if err != nil {
	//	return nil, err
	//}
	return buffer.Bytes(), nil
}

func (t *BlacklistChaincode) Fetch2(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	UserId := args[0]

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Failed getting metadata")
	}
	OrganizationId := string(OrganizationIdAsbytes)

	timeAsBytes, err := stub.GetState(LeaseStartPrefix + OrganizationId)
	if err != nil {
		return nil, errors.New("Failed getting lease start time")
	}

	leaseStartTime, err := time.Parse("2006-01-02 15:04:05", string(timeAsBytes))
	if err != nil {
		return nil, errors.New("Time conversion error")
	}

	// change to 5 seconds for test
	//if time.Now().Sub(leaseStartTime).Seconds() > 5 {
	//	return nil, errors.New("Lease expire error")
	//}
	if time.Now().Sub(leaseStartTime).Minutes() > 30 {
		return nil, errors.New("Lease expire error")
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

func (t *BlacklistChaincode) Fetch3(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	UserId := args[0]

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Failed getting metadata")
	}
	OrganizationId := string(OrganizationIdAsbytes)

	_, err = stub.GetState(LeaseEndPrefix + OrganizationId)
	if err != nil {
		return nil, errors.New("Lease may be expired")
	}
	fmt.Println("not expire======================================================================================= %v", LeaseEndPrefix + OrganizationId)
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
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
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
	readsBytes, err := stub.GetState(ReadsPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	buffer.WriteString("|")
	buffer.Write(readsBytes)
	sharesBytes, err := stub.GetState(SharesPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	buffer.WriteString("|")
	buffer.Write(sharesBytes)
	creditsBytes, err := stub.GetState(CreditsPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	buffer.WriteString("|")
	buffer.Write(creditsBytes)

	return buffer.Bytes(), nil
}

func main() {
	err := shim.Start(new(BlacklistChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

