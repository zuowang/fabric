/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"net"
	"testing"
	"time"

	//"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric/core/chaincode"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/hyperledger/fabric/core/chaincode/shim/crypto/attr"
	"github.com/hyperledger/fabric/core/container"
	"github.com/hyperledger/fabric/core/crypto"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/membersrvc/ca"
	pb "github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	//"reflect"
)

const (
	chaincodeStartupTimeoutDefault int = 30000
)

var (
	testLogger = logging.MustGetLogger("test")

	lis net.Listener

	administrator  crypto.Client
	idcUser       crypto.Client
	creditUser    crypto.Client
	microloanUser crypto.Client

	server *grpc.Server
	aca    *ca.ACA
	eca    *ca.ECA
	tca    *ca.TCA
	tlsca  *ca.TLSCA
)

func TestMain(m *testing.M) {
	removeFolders()
	setup()
	go initMembershipSrvc()

	fmt.Println("Wait for some secs for OBCCA")
	time.Sleep(2 * time.Second)

	go initVP()

	fmt.Println("Wait for some secs for VP")
	time.Sleep(2 * time.Second)

	go initBlacklistChaincode()

	fmt.Println("Wait for some secs for Chaincode")
	time.Sleep(2 * time.Second)

	if err := initClients(); err != nil {
		panic(err)
	}

	fmt.Println("Wait for 2 secs for chaincode to be started")
	time.Sleep(2 * time.Second)

	ret := m.Run()

	closeListenerAndSleep(lis)

	defer removeFolders()
	os.Exit(ret)
}

func TestFreeFetch(t *testing.T) {
	// Administrator deploy the chaicode
	adminCert, err := administrator.GetTCertificateHandlerNext("role")
	if err != nil {
		t.Fatal(err)
	}

	if err := deploy(adminCert, []string{"idc", "microloan", "credit"}); err != nil {
		t.Fatal(err)
	}

	// idcUser upload a blacklist
	if err := uploadBlacklist(idcUser, "idc", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploadBlacklist")

	blacklistBytes, err := fetchBlacklist(idcUser, "idc", []string{"370284197901130819"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("fetchBlacklist")

	idcAccountBytes, err := account(idcUser, "idc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("account")

	fmt.Printf("Fetch blacklist: %s\n", string(blacklistBytes))
	fmt.Printf("Query account: %s\n", string(idcAccountBytes))
}

func TestMultiItem(t *testing.T) {
	// Administrator deploy the chaicode
	adminCert, err := administrator.GetTCertificateHandlerNext("role")
	if err != nil {
		t.Fatal(err)
	}

	if err := deploy(adminCert, []string{"idc", "microloan", "credit"}); err != nil {
		t.Fatal(err)
	}

	// idcUser upload a blacklist
	if err := uploadBlacklist(idcUser, "idc", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	// microloanUser upload a blacklist
	if err := uploadBlacklist(microloanUser, "microloan", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "372922198012224773", "2016-07-12 16:37:21,2016-07-12", "230803197906010035", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploadBlacklist")

	blacklistBytes, err := fetchBlacklist(idcUser, "idc", []string{"370284197901130819"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("fetchBlacklist")

	idcAccountBytes, err := account(idcUser, "idc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("account")

	fmt.Printf("Fetch blacklist: %s\n", string(blacklistBytes))
	fmt.Printf("Query account: %s\n", string(idcAccountBytes))
}

func TestDeleteItem(t *testing.T) {
	// Administrator deploy the chaicode
	adminCert, err := administrator.GetTCertificateHandlerNext("role")
	if err != nil {
		t.Fatal(err)
	}

	if err := deploy(adminCert, []string{"idc", "microloan", "credit"}); err != nil {
		t.Fatal(err)
	}

	// idcUser upload a blacklist
	if err := uploadBlacklist(idcUser, "idc", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	// microloanUser upload a blacklist
	if err := uploadBlacklist(microloanUser, "microloan", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "372922198012224773", "2016-07-12 16:37:21,2016-07-12", "230803197906010035", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploadBlacklist")

	blacklistBytes, err := fetchBlacklist(idcUser, "idc", []string{"370284197901130819"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("fetchBlacklist")

	idcAccountBytes, err := account(idcUser, "idc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("account")

	// idcUser delete 370284197901130819 from blacklist
	if err := deleteBlacklist(microloanUser, "microloan", []string{"370284197901130819"}); err != nil {
		t.Fatal(err)
	}

	blacklistBytes2, err := fetchBlacklist(idcUser, "idc", []string{"370284197901130819"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Fetch blacklist: %s\n", string(blacklistBytes))
	fmt.Printf("Query account: %s\n", string(idcAccountBytes))
	fmt.Printf("Fetch blacklist: %s\n", string(blacklistBytes2))
}

func TestPayPerQuery(t *testing.T) {
	// Administrator deploy the chaicode
	adminCert, err := administrator.GetTCertificateHandlerNext("role")
	if err != nil {
		t.Fatal(err)
	}

	if err := deploy(adminCert, []string{"idc", "microloan", "credit"}); err != nil {
		t.Fatal(err)
	}

	// idcUser upload a blacklist
	if err := uploadBlacklist(idcUser, "idc", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	// microloanUser upload a blacklist
	if err := uploadBlacklist(microloanUser, "microloan", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "372922198012224773", "2016-07-12 16:37:21,2016-07-12", "230803197906010035", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploadBlacklist")

	idcAccountBytes, err := account(idcUser, "idc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("account")

	blacklistBytes, err := readBlacklist(idcUser, "idc", []string{"370284197901130819"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Query account: %s\n", string(idcAccountBytes))
	fmt.Printf("Read blacklist: %s\n", string(blacklistBytes))
}

func TestLease(t *testing.T) {
	// Administrator deploy the chaicode
	adminCert, err := administrator.GetTCertificateHandlerNext("role")
	if err != nil {
		t.Fatal(err)
	}

	if err := deploy(adminCert, []string{"idc", "microloan", "credit"}); err != nil {
		t.Fatal(err)
	}

	// idcUser upload a blacklist
	if err := uploadBlacklist(idcUser, "idc", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	// microloanUser upload a blacklist
	if err := uploadBlacklist(microloanUser, "microloan", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "372922198012224773", "2016-07-12 16:37:21,2016-07-12", "230803197906010035", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploadBlacklist")

	if err := extendLease(idcUser, "idc", []string{}); err != nil {
		t.Fatal(err)
	}

	blacklistBytes, err := fetch2Blacklist(idcUser, "idc", []string{"370284197901130819"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("fetch2Blacklist")

	idcAccountBytes, err := account(idcUser, "idc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Fetch2 blacklist: %s\n", string(blacklistBytes))
	fmt.Printf("Query account: %s\n", string(idcAccountBytes))
}

func TestLeaseExpire(t *testing.T) {
	// Administrator deploy the chaicode
	adminCert, err := administrator.GetTCertificateHandlerNext("role")
	if err != nil {
		t.Fatal(err)
	}

	if err := deploy(adminCert, []string{"idc", "microloan", "credit"}); err != nil {
		t.Fatal(err)
	}

	// idcUser upload a blacklist
	if err := uploadBlacklist(idcUser, "idc", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	// microloanUser upload a blacklist
	if err := uploadBlacklist(microloanUser, "microloan", []string{"370284197901130819", "2016-07-12 16:37:21,2016-07-12", "372922198012224773", "2016-07-12 16:37:21,2016-07-12", "230803197906010035", "2016-07-12 16:37:21,2016-07-12"}); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploadBlacklist")

	if err := extendLease(idcUser, "idc", []string{}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(5*1000*1000*1000))

	_, err = fetch2Blacklist(idcUser, "idc", []string{"370284197901130819"})
	if err == nil {
		t.Fatal(err)
	}

	fmt.Println("fetch2Blacklist")

	idcAccountBytes, err := account(idcUser, "idc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Query account: %s\n", string(idcAccountBytes))
}

func deploy(admCert crypto.CertificateHandler,args []string) error {
	// Prepare the spec. The metadata includes the role of the users allowed to assign assets
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              &pb.ChaincodeInput{Function: "init", Args: args},
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	// First build and get the deployment spec
	var ctx = context.Background()
	chaincodeDeploymentSpec, err := getDeploymentSpec(ctx, spec)
	if err != nil {
		return err
	}

	tid := chaincodeDeploymentSpec.ChaincodeSpec.ChaincodeID.Name

	// Now create the Transactions message and send to Peer.
	transaction, err := administrator.NewChaincodeDeployTransaction(chaincodeDeploymentSpec, tid)
	if err != nil {
		return fmt.Errorf("Error deploying chaincode: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	_, _, err = chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return fmt.Errorf("Error deploying chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return err
}

func uploadBlacklist(client crypto.Client, orgName string, args []string) error {
	// Get a transaction handler to be used to submit the execute transaction
	// and bind the chaincode access control logic using the binding
	submittingCertHandler, err := client.GetTCertificateHandlerNext("role")
	if err != nil {
		return err
	}
	txHandler, err := submittingCertHandler.GetTransactionHandler()
	if err != nil {
		return err
	}

	chaincodeInput := &pb.ChaincodeInput{Function: "write", Args: args}

	// Prepare spec and submit
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              chaincodeInput,
		Metadata:             []byte(orgName),
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	var ctx = context.Background()
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	tid := chaincodeInvocationSpec.ChaincodeSpec.ChaincodeID.Name


	// Now create the Transactions message and send to Peer.
	transaction, err := txHandler.NewChaincodeExecute(chaincodeInvocationSpec, tid)//client.NewChaincodeExecute(chaincodeInvocationSpec, tid, []string{"role"}...) //
	if err != nil {
		return fmt.Errorf("Error new transaction: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	_, _, err = chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return fmt.Errorf("Error invoking chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return err
}

func fetchBlacklist(client crypto.Client, orgName string, args []string) ([]byte, error) {
	// Get a transaction handler to be used to submit the execute transaction
	// and bind the chaincode access control logic using the binding
	submittingCertHandler, err := client.GetTCertificateHandlerNext("role")
	if err != nil {
		return nil, err
	}
	txHandler, err := submittingCertHandler.GetTransactionHandler()
	if err != nil {
		return nil, err
	}

	chaincodeInput := &pb.ChaincodeInput{Function: "fetch", Args: args}

	// Prepare spec and submit
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              chaincodeInput,
		Metadata:             []byte(orgName),
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	var ctx = context.Background()
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	tid := chaincodeInvocationSpec.ChaincodeSpec.ChaincodeID.Name

	// Now create the Transactions message and send to Peer.
	transaction, err := txHandler.NewChaincodeQuery(chaincodeInvocationSpec, tid)
	if err != nil {
		return nil, fmt.Errorf("Error new transaction: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	valBytes, _, err := chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return nil, fmt.Errorf("Error query chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return valBytes, err
}

func account(client crypto.Client, orgName string, args []string) ([]byte, error) {
	// Get a transaction handler to be used to submit the execute transaction
	// and bind the chaincode access control logic using the binding
	submittingCertHandler, err := client.GetTCertificateHandlerNext("role")
	if err != nil {
		return nil, err
	}
	txHandler, err := submittingCertHandler.GetTransactionHandler()
	if err != nil {
		return nil, err
	}

	chaincodeInput := &pb.ChaincodeInput{Function: "account", Args: args}

	// Prepare spec and submit
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              chaincodeInput,
		Metadata:             []byte(orgName),
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	var ctx = context.Background()
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	tid := chaincodeInvocationSpec.ChaincodeSpec.ChaincodeID.Name

	// Now create the Transactions message and send to Peer.
	transaction, err := txHandler.NewChaincodeQuery(chaincodeInvocationSpec, tid)
	if err != nil {
		return nil, fmt.Errorf("Error new transaction: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	valBytes, _, err := chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return nil, fmt.Errorf("Error query chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return valBytes, err
}

func deleteBlacklist(client crypto.Client, orgName string, args []string) error {
	// Get a transaction handler to be used to submit the execute transaction
	// and bind the chaincode access control logic using the binding
	submittingCertHandler, err := client.GetTCertificateHandlerNext("role")
	if err != nil {
		return err
	}
	txHandler, err := submittingCertHandler.GetTransactionHandler()
	if err != nil {
		return err
	}

	chaincodeInput := &pb.ChaincodeInput{Function: "delete", Args: args}

	// Prepare spec and submit
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              chaincodeInput,
		Metadata:             []byte(orgName),
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	var ctx = context.Background()
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	tid := chaincodeInvocationSpec.ChaincodeSpec.ChaincodeID.Name

	// Now create the Transactions message and send to Peer.
	transaction, err := txHandler.NewChaincodeExecute(chaincodeInvocationSpec, tid)
	if err != nil {
		return fmt.Errorf("Error new transaction: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	_, _, err = chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return fmt.Errorf("Error invoking chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return err
}

func readBlacklist(client crypto.Client, orgName string, args []string) ([]byte, error) {
	// Get a transaction handler to be used to submit the execute transaction
	// and bind the chaincode access control logic using the binding
	submittingCertHandler, err := client.GetTCertificateHandlerNext("role")
	if err != nil {
		return nil, err
	}
	txHandler, err := submittingCertHandler.GetTransactionHandler()
	if err != nil {
		return nil, err
	}

	chaincodeInput := &pb.ChaincodeInput{Function: "read", Args: args}

	// Prepare spec and submit
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              chaincodeInput,
		Metadata:             []byte(orgName),
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	var ctx = context.Background()
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	tid := chaincodeInvocationSpec.ChaincodeSpec.ChaincodeID.Name

	// Now create the Transactions message and send to Peer.
	transaction, err := txHandler.NewChaincodeExecute(chaincodeInvocationSpec, tid)
	if err != nil {
		return nil, fmt.Errorf("Error new transaction: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	valBytes, _, err := chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return nil, fmt.Errorf("Error invoking chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return valBytes, err
}

func extendLease(client crypto.Client, orgName string, args []string) error {
	// Get a transaction handler to be used to submit the execute transaction
	// and bind the chaincode access control logic using the binding
	submittingCertHandler, err := client.GetTCertificateHandlerNext("role")
	if err != nil {
		return err
	}
	txHandler, err := submittingCertHandler.GetTransactionHandler()
	if err != nil {
		return err
	}

	chaincodeInput := &pb.ChaincodeInput{Function: "lease", Args: args}

	// Prepare spec and submit
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              chaincodeInput,
		Metadata:             []byte(orgName),
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	var ctx = context.Background()
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	tid := chaincodeInvocationSpec.ChaincodeSpec.ChaincodeID.Name


	// Now create the Transactions message and send to Peer.
	transaction, err := txHandler.NewChaincodeExecute(chaincodeInvocationSpec, tid)//client.NewChaincodeExecute(chaincodeInvocationSpec, tid, []string{"role"}...) //
	if err != nil {
		return fmt.Errorf("Error new transaction: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	_, _, err = chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return fmt.Errorf("Error invoking chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return err
}

func fetch2Blacklist(client crypto.Client, orgName string, args []string) ([]byte, error) {
	// Get a transaction handler to be used to submit the execute transaction
	// and bind the chaincode access control logic using the binding
	submittingCertHandler, err := client.GetTCertificateHandlerNext("role")
	if err != nil {
		return nil, err
	}
	txHandler, err := submittingCertHandler.GetTransactionHandler()
	if err != nil {
		return nil, err
	}

	chaincodeInput := &pb.ChaincodeInput{Function: "fetch2", Args: args}

	// Prepare spec and submit
	spec := &pb.ChaincodeSpec{
		Type:                 1,
		ChaincodeID:          &pb.ChaincodeID{Name: "mycc"},
		CtorMsg:              chaincodeInput,
		Metadata:             []byte(orgName),
		ConfidentialityLevel: pb.ConfidentialityLevel_PUBLIC,
	}

	var ctx = context.Background()
	chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	tid := chaincodeInvocationSpec.ChaincodeSpec.ChaincodeID.Name

	// Now create the Transactions message and send to Peer.
	transaction, err := txHandler.NewChaincodeQuery(chaincodeInvocationSpec, tid)
	if err != nil {
		return nil, fmt.Errorf("Error new transaction: %s ", err)
	}

	ledger, err := ledger.GetLedger()
	ledger.BeginTxBatch("1")
	valBytes, _, err := chaincode.Execute(ctx, chaincode.GetChain(chaincode.DefaultChain), transaction)
	if err != nil {
		return nil, fmt.Errorf("Error query chaincode: %s", err)
	}
	ledger.CommitTxBatch("1", []*pb.Transaction{transaction}, nil, nil)

	return valBytes, err
}

func setup() {
	// Conf
	viper.SetConfigName("blacklist") // name of config file (without extension)
	viper.AddConfigPath(".")     // path to look for the config file in
	err := viper.ReadInConfig()  // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file [%s] \n", err))
	}

	// Logging
	var formatter = logging.MustStringFormatter(
		`%{color}[%{module}] %{shortfunc} [%{shortfile}] -> %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	logging.SetFormatter(formatter)

	logging.SetLevel(logging.DEBUG, "peer")
	logging.SetLevel(logging.DEBUG, "chaincode")
	logging.SetLevel(logging.DEBUG, "cryptochain")

	// Init the crypto layer
	if err := crypto.Init(); err != nil {
		panic(fmt.Errorf("Failed initializing the crypto layer [%s]", err))
	}

	removeFolders()
}

func initMembershipSrvc() {
	ca.LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr, os.Stdout)

	aca = ca.NewACA()
	eca = ca.NewECA()
	tca = ca.NewTCA(eca)
	tlsca = ca.NewTLSCA(eca)

	var opts []grpc.ServerOption
	if viper.GetBool("peer.pki.tls.enabled") {
		// TLS configuration
		creds, err := credentials.NewServerTLSFromFile(
			filepath.Join(viper.GetString("server.rootpath"), "tlsca.cert"),
			filepath.Join(viper.GetString("server.rootpath"), "tlsca.priv"),
		)
		if err != nil {
			panic("Failed creating credentials for Membersrvc: " + err.Error())
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	fmt.Printf("open socket...\n")
	sockp, err := net.Listen("tcp", viper.GetString("server.port"))
	if err != nil {
		panic("Cannot open port: " + err.Error())
	}
	fmt.Printf("open socket...done\n")

	server = grpc.NewServer(opts...)

	aca.Start(server)
	eca.Start(server)
	tca.Start(server)
	tlsca.Start(server)

	fmt.Printf("start serving...\n")
	server.Serve(sockp)
}

func initVP() {
	var opts []grpc.ServerOption
	if viper.GetBool("peer.tls.enabled") {
		creds, err := credentials.NewServerTLSFromFile(viper.GetString("peer.tls.cert.file"), viper.GetString("peer.tls.key.file"))
		if err != nil {
			grpclog.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)

	//lis, err := net.Listen("tcp", viper.GetString("peer.address"))

	//use a different address than what we usually use for "peer"
	//we override the peerAddress set in chaincode_support.go
	peerAddress := "0.0.0.0:40404"
	var err error
	lis, err = net.Listen("tcp", peerAddress)
	if err != nil {
		return
	}

	getPeerEndpoint := func() (*pb.PeerEndpoint, error) {
		return &pb.PeerEndpoint{ID: &pb.PeerID{Name: "testpeer"}, Address: peerAddress}, nil
	}

	ccStartupTimeout := time.Duration(chaincodeStartupTimeoutDefault) * time.Millisecond
	userRunsCC := true

	// Install security object for peer
	var secHelper crypto.Peer
	if viper.GetBool("security.enabled") {
		enrollID := viper.GetString("security.enrollID")
		enrollSecret := viper.GetString("security.enrollSecret")
		var err error

		if viper.GetBool("peer.validator.enabled") {
			testLogger.Debugf("Registering validator with enroll ID: %s", enrollID)
			if err = crypto.RegisterValidator(enrollID, nil, enrollID, enrollSecret); nil != err {
				panic(err)
			}
			testLogger.Debugf("Initializing validator with enroll ID: %s", enrollID)
			secHelper, err = crypto.InitValidator(enrollID, nil)
			if nil != err {
				panic(err)
			}
		} else {
			testLogger.Debugf("Registering non-validator with enroll ID: %s", enrollID)
			if err = crypto.RegisterPeer(enrollID, nil, enrollID, enrollSecret); nil != err {
				panic(err)
			}
			testLogger.Debugf("Initializing non-validator with enroll ID: %s", enrollID)
			secHelper, err = crypto.InitPeer(enrollID, nil)
			if nil != err {
				panic(err)
			}
		}
	}

	pb.RegisterChaincodeSupportServer(grpcServer,
		chaincode.NewChaincodeSupport(chaincode.DefaultChain, getPeerEndpoint, userRunsCC,
			ccStartupTimeout, secHelper))

	grpcServer.Serve(lis)
}

func initBlacklistChaincode() {
	err := shim.Start(new(BlacklistChaincode))
	if err != nil {
		panic(err)
	}
}

func initClients() error {
	// Administrator
	if err := crypto.RegisterClient("admin", nil, "admin", "6avZQLwcUe9b"); err != nil {
		return err
	}
	var err error
	administrator, err = crypto.InitClient("admin", nil)
	if err != nil {
		return err
	}

	// idc_user
	if err := crypto.RegisterClient("idcUser", nil, "idcUser", "6avZQLwcUe9b"); err != nil {
		return err
	}
	idcUser, err = crypto.InitClient("idcUser", nil)
	if err != nil {
		return err
	}

	// credit_user
	if err := crypto.RegisterClient("creditUser", nil, "creditUser", "DRJ23pEQl16a"); err != nil {
		return err
	}
	creditUser, err = crypto.InitClient("creditUser", nil)
	if err != nil {
		return err
	}

	// microloan_user
	if err := crypto.RegisterClient("microloanUser", nil, "microloanUser", "NPKYL39uKbkj"); err != nil {
		return err
	}
	microloanUser, err = crypto.InitClient("microloanUser", nil)
	if err != nil {
		return err
	}

	return nil
}

func closeListenerAndSleep(l net.Listener) {
	l.Close()
	time.Sleep(2 * time.Second)
}

func getDeploymentSpec(context context.Context, spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	fmt.Printf("getting deployment spec for chaincode spec: %v\n", spec)
	var codePackageBytes []byte
	//if we have a name, we don't need to deploy (we are in userRunsCC mode)
	if spec.ChaincodeID.Name == "" {
		var err error
		codePackageBytes, err = container.GetChaincodePackageBytes(spec)
		if err != nil {
			return nil, err
		}
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return chaincodeDeploymentSpec, nil
}

func removeFolders() {
	if err := os.RemoveAll(viper.GetString("peer.fileSystemPath")); err != nil {
		fmt.Printf("Failed removing [%s] [%s]\n", "hyperledger", err)
	}
}
