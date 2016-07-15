#!/bin/sh bash

build/bin/membersrvc

docker run --rm -it -e CORE_VM_ENDPOINT=http://172.17.0.1:2375 -e CORE_PEER_ID=vp0 -e CORE_PEER_ADDRESSAUTODETECT=true -e CORE_SECURITY_ENABLED=true -e CORE_SECURITY_PRIVACY=true -e CORE_PEER_PKI_ECA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TCA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TLSCA_PADDR=10.199.90.105:50051 -e CORE_SECURITY_ENROLLID=test_vp0 -e CORE_SECURITY_ENROLLSECRET=MwYpmSRjupbT -e CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN=pbft -e CORE_PBFT_GENERAL_MODE=batch -e CORE_LOGGING_LEVEL=CRITICAL -e CORE_PEER_PROFILE_ENABLED=true -p 30303:30303 -p 31315:31315 -p 5000:5000 -e CORE_SECURITY_TCERT_BATCH_SIZE=1 hyperledger/fabric-peer peer node start

docker run --rm -it -e CORE_VM_ENDPOINT=http://172.17.0.1:2375 -e CORE_PEER_ID=vp1 -e CORE_PEER_ADDRESSAUTODETECT=true -e CORE_PEER_DISCOVERY_ROOTNODE=172.17.0.4:30303 -e CORE_SECURITY_ENABLED=true -e CORE_SECURITY_PRIVACY=true -e CORE_PEER_PKI_ECA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TCA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TLSCA_PADDR=10.199.90.105:50051 -e CORE_SECURITY_ENROLLID=test_vp1 -e CORE_SECURITY_ENROLLSECRET=5wgHK9qqYaPy -e CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN=pbft -e CORE_PBFT_GENERAL_MODE=batch -e CORE_LOGGING_LEVEL=CRITICAL -e CORE_SECURITY_TCERT_BATCH_SIZE=1  hyperledger/fabric-peer peer node start

docker run --rm -it -e CORE_VM_ENDPOINT=http://172.17.0.1:2375 -e CORE_PEER_ID=vp2 -e CORE_PEER_ADDRESSAUTODETECT=true -e CORE_PEER_DISCOVERY_ROOTNODE=172.17.0.4:30303 -e CORE_SECURITY_ENABLED=true -e CORE_SECURITY_PRIVACY=true -e CORE_PEER_PKI_ECA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TCA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TLSCA_PADDR=10.199.90.105:50051 -e CORE_SECURITY_ENROLLID=test_vp2 -e CORE_SECURITY_ENROLLSECRET=vQelbRvja7cJ -e CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN=pbft -e CORE_PBFT_GENERAL_MODE=batch -e CORE_LOGGING_LEVEL=CRITICAL -e CORE_SECURITY_TCERT_BATCH_SIZE=1  hyperledger/fabric-peer peer node start

docker run --rm -it -e CORE_VM_ENDPOINT=http://172.17.0.1:2375 -e CORE_PEER_ID=vp3 -e CORE_PEER_ADDRESSAUTODETECT=true -e CORE_PEER_DISCOVERY_ROOTNODE=172.17.0.4:30303 -e CORE_SECURITY_ENABLED=true -e CORE_SECURITY_PRIVACY=true -e CORE_PEER_PKI_ECA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TCA_PADDR=10.199.90.105:50051 -e CORE_PEER_PKI_TLSCA_PADDR=10.199.90.105:50051 -e CORE_SECURITY_ENROLLID=test_vp3 -e CORE_SECURITY_ENROLLSECRET=9LKqKH5peurL -e CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN=pbft -e CORE_PBFT_GENERAL_MODE=batch -e CORE_LOGGING_LEVEL=CRITICAL -e CORE_SECURITY_TCERT_BATCH_SIZE=1  hyperledger/fabric-peer peer node start



CORE_PEER_ADDRESS=172.17.0.4:30303 build/bin/peer network login jim -p 6avZQLwcUe9b

CORE_PEER_ADDRESS=172.17.0.4:30303 CORE_SECURITY_ENABLED=true CORE_SECURITY_PRIVACY=true build/bin/peer chaincode deploy -u jim -p github.com/hyperledger/fabric/examples/chaincode/go/chaincode_blacklist -c '{"Function":"init", "Args": ["jim","diego", "binhn"]}'

CORE_PEER_ADDRESS=172.17.0.4:30303 CORE_SECURITY_ENABLED=true CORE_SECURITY_PRIVACY=true build/bin/peer chaincode invoke -u jim -n 12f7134e92656315778e0955705c76f0661d0fa7a2556b434ff6255a51fb575dbd0264e2dcbbf175e13357c47b0c623d110b67f5d498a4897280988aefa5de10 -c '{"Function":"write", "Args": ["370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"]}' -a '["role"]'

CORE_PEER_ADDRESS=172.17.0.4:30303 CORE_SECURITY_ENABLED=true CORE_SECURITY_PRIVACY=true build/bin/peer chaincode invoke -u jim -l golang -n 12f7134e92656315778e0955705c76f0661d0fa7a2556b434ff6255a51fb575dbd0264e2dcbbf175e13357c47b0c623d110b67f5d498a4897280988aefa5de10 -c '{"Function": "read", "Args": ["370284197901130819"]}' -a '["role"]'

CORE_PEER_ADDRESS=172.17.0.4:30303 CORE_SECURITY_ENABLED=true CORE_SECURITY_PRIVACY=true build/bin/peer chaincode query -u jim -l golang -n 12f7134e92656315778e0955705c76f0661d0fa7a2556b434ff6255a51fb575dbd0264e2dcbbf175e13357c47b0c623d110b67f5d498a4897280988aefa5de10 -c '{"Function": "fetch", "Args": ["370284197901130819"]}' -a '["role"]'



curl -X GET --header "Accept: application/json" "http://localhost:5000/registrar/jim/ecert"

curl -X POST --header "Content-Type: application/x-www-form-urlencoded" --header "Accept: application/json" -d '{  "enrollId": "jim",  "enrollSecret": "6avZQLwcUe9b" }' "http://localhost:5000/registrar"
curl -X POST --header "Content-Type: application/x-www-form-urlencoded" --header "Accept: application/json" -d '{  "enrollId": "bob",  "enrollSecret": "NOE63pEQbL25" }' "http://localhost:5000/registrar"

curl -i -X POST -H "Content-Type: application/json" http://localhost:5000/chaincode -d '{ "jsonrpc": "2.0", "method": "deploy", "params": { "type": 1, "chaincodeID":{ "path":"github.com/hyperledger/fabric/examples/chaincode/go/chaincode_blacklist" },"ctorMsg": { "function":"init", "args":["jim","diego", "binhn"] },"secureContext":"jim", "confidentialityLevel":1, "metadata":"aWRj", "attributes":["role"] }, "id": 5}'
curl -i -X POST -H "Content-Type: application/json" http://localhost:5000/chaincode -d '{ "jsonrpc": "2.0", "method": "deploy", "params": { "type": 1, "chaincodeID":{ "path":"github.com/hyperledger/fabric/examples/chaincode/go/chaincode_blacklist" },"ctorMsg": { "function":"init", "args":["jim","bob", "binhn"] },"secureContext":"bob", "confidentialityLevel":1, "metadata":"Ym9i", "attributes":["role"] }, "id": 5}'

curl -i -X POST -H "Content-Type: application/json" http://localhost:5000/chaincode -d '{ "jsonrpc": "2.0", "method": "invoke", "params": { "type": 1, "chaincodeID":{ "name":"12f7134e92656315778e0955705c76f0661d0fa7a2556b434ff6255a51fb575dbd0264e2dcbbf175e13357c47b0c623d110b67f5d498a4897280988aefa5de10" },"ctorMsg": { "function":"write", "args":["370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"] }, "secureContext":"jim", "confidentialityLevel":1, "metadata":"amlt", "attributes":["role"]  }, "id": 5}'
curl -i -X POST -H "Content-Type: application/json" http://localhost:5000/chaincode -d '{ "jsonrpc": "2.0", "method": "invoke", "params": { "type": 1, "chaincodeID":{ "name":"b309bbbb319bdcb6498440986b6a13aef04c3cb562d4a2204c79b7745a4b76ea5abf7cb6e5c93aff596a5576923c67e1d45155cd75e3b48771e884fc94e4d1a9" },"ctorMsg": { "function":"write", "args":["370284197901130819", "2016-07-12 16:37:21,2016-07-12", "210905197807210546", "2016-07-12 16:37:21,2016-07-12", "370205197405213513", "2016-07-12 16:37:21,2016-07-12"] }, "secureContext":"bob", "confidentialityLevel":1, "metadata":"Ym9i", "attributes":["role"]  }, "id": 5}'

curl -i -X POST -H "Content-Type: application/json" http://localhost:5000/chaincode -d '{ "jsonrpc": "2.0", "method": "invoke", "params": { "type": 1, "chaincodeID":{ "name":"30d78c2e5d560e9118e2db0c3577cfcbee5f88592ce4c7e24c748ff994314602c82ea75dfbe762ea2896529dabbfc5d302d05d1ac8096bb1f6a7c8f7de306316" },"ctorMsg": { "function":"write", "args":["370284197901130819", "2016-07-12 16:37:21,2016-07-12", "372922198012224773", "2016-07-12 16:37:21,2016-07-12", "230803197906010035", "2016-07-12 16:37:21,2016-07-12"] }, "secureContext":"jim", "confidentialityLevel":1, "metadata":"aWRj", "attributes":["role"] }, "id": 5}'

curl -i -X POST -H "Content-Type: application/json" http://localhost:5000/chaincode -d '{ "jsonrpc": "2.0", "method": "query", "params": { "type": 1, "chaincodeID":{ "name":"12f7134e92656315778e0955705c76f0661d0fa7a2556b434ff6255a51fb575dbd0264e2dcbbf175e13357c47b0c623d110b67f5d498a4897280988aefa5de10" },"ctorMsg": { "function":"fetch", "args":["370284197901130819"] }, "secureContext":"jim", "confidentialityLevel":1, "metadata":"aWRj", "attributes":["role"] }, "id": 5}'








