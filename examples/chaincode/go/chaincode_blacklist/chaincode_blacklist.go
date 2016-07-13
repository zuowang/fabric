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
	"strings"
)

var chaincodeLogger = logging.MustGetLogger("blacklist")

// BlacklistChaincode implements the insert into and query of blacklist providing user with a "Organization" role.
type BlacklistChaincode struct {
}

var WritesPrefix = "#w"
var ReadsPrefix = "#r"
var SharesPrefix = "#s"
var CreditsPrefix = "#c"
var SharesSuffix = "#"

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
	//callerRole, err := stub.ReadCertAttribute("role")
	//if err != nil {
	//	chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
	//	return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	//}
	//
	//caller := string(callerRole)
	//if (caller != "Organization") {
	//	chaincodeLogger.Errorf("Failed validating caller role")
	//	return nil, fmt.Errorf("Failed validating caller role.")
	//}
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

	//OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	//if err != nil {
	//	return nil, err
	//}
	//OrganizationId := string(OrganizationIdAsbytes)
	OrganizationId := ""
	//OldWritesAsbytes, err := stub.GetState(WritesPrefix + OrganizationId)
	//if err != nil {
	//	return nil, err
	//}
	//OldWrites, err := strconv.Atoi(string(OldWritesAsbytes))
	//if err != nil {
	//	return nil, err
	//}
	OldWrites := 0
	var err error
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

		chaincodeLogger.Infof("PutState [%v]", args[i - 1] + OrganizationId + SharesSuffix)
		err = stub.PutState(args[i - 1] + OrganizationId + SharesSuffix, []byte(strconv.Itoa(0)))
		if err != nil {
			return nil, err
		}
		OldWrites++
	}
	//err = stub.PutState(OrganizationId, []byte(strconv.Itoa(OldWrites)))
	//if err != nil {
	//	return nil, err
	//}

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
		err = stub.DelState(UserId + OrganizationId + SharesSuffix)
		if err != nil {
			return nil, err
		}
		OldWrites--
	}

	err = stub.PutState(OrganizationId, []byte(strconv.Itoa(OldWrites)))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *BlacklistChaincode) Read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	//callerRole, err := stub.ReadCertAttribute("role")
	//if err != nil {
	//	chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
	//	return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	//}
	//
	//caller := string(callerRole)
	//if (caller != "Organization") {
	//	chaincodeLogger.Errorf("Failed validating caller role")
	//	return nil, fmt.Errorf("Failed validating caller role.")
	//}

	UserId := args[0]
	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, errors.New("Failed getting metadata")
	}
	OrganizationId := string(OrganizationIdAsbytes)

	iter, err := stub.RangeQueryState(UserId+"1", UserId+":")
	if err != nil {
		return nil, fmt.Errorf("Error fetching blacklist: [%v]", err)
	}
	defer iter.Close()

	for iter.HasNext() {
		_, valBytes, err := iter.Next()
		if err != nil {
			return nil, err
		}
		val := strings.Split(string(valBytes), ",")
		// Read counter plus one if read other's blacklist
		if (val[1] != OrganizationId) {
			OldReadsAsbytes, err := stub.GetState(ReadsPrefix + OrganizationId)
			if err != nil {
				return nil, err
			}
			OldReads, err := strconv.Atoi(string(OldReadsAsbytes))
			if err != nil {
				return nil, err
			}
			err = stub.PutState(OrganizationId, []byte(strconv.Itoa(OldReads + 1)))
			if err != nil {
				return nil, err
			}

			OldSharesAsbytes, err := stub.GetState(SharesPrefix + val[1])
			if err != nil {
				return nil, err
			}
			OldShares, err := strconv.Atoi(string(OldSharesAsbytes))
			if err != nil {
				return nil, err
			}
			err = stub.PutState(SharesPrefix + val[1], []byte(strconv.Itoa(OldShares + 1)))
			if err != nil {
				return nil, err
			}

			OldSharesAsbytes, err = stub.GetState(UserId + val[1] + SharesSuffix)
			if err != nil {
				return nil, err
			}
			OldShares, err = strconv.Atoi(string(OldSharesAsbytes))
			if err != nil {
				return nil, err
			}
			err = stub.PutState(UserId + val[1] + SharesSuffix, []byte(strconv.Itoa(OldShares + 1)))
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

func (t *BlacklistChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
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

	UserId := args[0]
	//iter, err := stub.RangeQueryState(UserId, UserId+":")
	//if err != nil {
	//	return nil, fmt.Errorf("Error fetching blacklist: [%v]", err)
	//}
	//defer iter.Close()
	//
	//var resultBytes []byte
	//resultBytes = nil
	//for iter.HasNext() {
	//	_, valBytes, err := iter.Next()
	//	if err != nil {
	//		return nil, err
	//	}
	//	resultBytes = valBytes
	//}
	//
	//if resultBytes == nil {
	//	return nil, fmt.Errorf("Error Nil for [%v]", UserId)
	//}
	//OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	//if err != nil {
	//	return nil, errors.New("Failed getting metadata")
	//}
	//OrganizationId := string(OrganizationIdAsbytes)
	OrganizationId := ""
	chaincodeLogger.Infof("GetState [%v]", UserId + OrganizationId)
	resultBytes, err := stub.GetState(UserId + OrganizationId)
	if err != nil {
		return nil, err
	}
	return resultBytes, nil
}

func (t *BlacklistChaincode) Account(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	//callerRole, err := stub.ReadCertAttribute("role")
	//if err != nil {
	//	chaincodeLogger.Errorf("Error reading attribute: [%v]", err)
	//	return nil, fmt.Errorf("Failed fetching caller role. Error was [%v]", err)
	//}
	//
	//caller := string(callerRole)
	//if (caller != "Organization") {
	//	chaincodeLogger.Errorf("Failed validating caller role")
	//	return nil, fmt.Errorf("Failed validating caller role.")
	//}

	OrganizationIdAsbytes, err := stub.GetCallerMetadata()
	if err != nil {
		return nil, err
	}
	OrganizationId := string(OrganizationIdAsbytes)
	writesBytes, err := stub.GetState(WritesPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	_, err = stub.GetState(ReadsPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}
	_, err = stub.GetState(SharesPrefix + OrganizationId)
	if err != nil {
		return nil, err
	}

	return writesBytes, nil
}

func main() {
	err := shim.Start(new(BlacklistChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

