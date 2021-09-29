/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	yobaNetwork, err := gw.GetNetwork("yobachannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	basicContract := yobaNetwork.GetContract("basic")
	log.Println("First call to 'yobachannel/basic'")
	result, err := basicContract.EvaluateTransaction(context.Background(), "GetAllAssets")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println("yobachannel/basic", string(result))

	// second transaction to different chaincode
	tr, err := basicContract.CreateTransaction(
		"ReadAsset",
		gateway.InsecureWithChannel("bobachannel"),
		gateway.InsecureWithChaincode("ledqueries"),
		//gateway.InsecureWithWalletIdentity()
	)
	if err != nil {
		log.Fatalf("Failed to create transaction: %v", err)
	}

	log.Println("Second call to 'bobachannel/ledqueries'")
	r, err := tr.Evaluate(context.Background(), "asset1337")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println("bobachannel/ledqueries", string(r))

	// -- second call manual
	// bobaNetwork, err := gw.GetNetwork("bobachannel")
	// if err != nil {
	// 	log.Fatalf("Failed to get network: %v", err)
	// }
	// queryContract := bobaNetwork.GetContract("ledqueries")

	// r, err := queryContract.EvaluateTransaction(context.Background(), "ReadAsset", "asset1337")
	// if err != nil {
	// 	log.Fatalf("Failed to evaluate transaction: %v", err)
	// }
	// log.Println("bobachannel/ledqueries", string(r))

	log.Println("============ application-golang ends ============")
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
