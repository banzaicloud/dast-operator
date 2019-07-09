/*
Copyright 2019 Banzai Cloud.

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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/zaproxy/zap-api-go/zap"
)

var zapAddr string
var target string
var apiKey string
var serve bool

var scannerCmd = &cobra.Command{
	Use:   "scanner",
	Short: "Scanner application using Zap",
	Run: func(cmd *cobra.Command, args []string) {
		scanner()
	},
}

func init() {
	scannerCmd.Flags().StringVarP(&zapAddr, "zap-proxy", "p", "http://127.0.0.1:8080", "Zap proxy address")
	scannerCmd.Flags().StringVarP(&target, "target", "t", "http://127.0.0.1:8090/target", "Target address")
	scannerCmd.Flags().StringVarP(&apiKey, "apikey", "a", os.Getenv("ZAPAPIKEY"), "Zap api key")
	scannerCmd.Flags().BoolVarP(&serve, "serve", "s", false, "serve results")
	rootCmd.AddCommand(scannerCmd)
}

func scanner() {
	cfg := &zap.Config{
		Proxy:  zapAddr,
		APIKey: apiKey,
	}
	client, err := zap.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Start spidering the target
	fmt.Println("Spider : " + target)
	resp, err := client.Spider().Scan(target, "", "", "", "")
	if err != nil {
		log.Fatal(err)
	}

	// The scan now returns a scan id to support concurrent scanning
	scanid := resp["scan"].(string)
	for {
		time.Sleep(1000 * time.Millisecond)
		resp, _ = client.Spider().Status(scanid)
		progress, _ := strconv.Atoi(resp["status"].(string))
		if progress >= 100 {
			break
		}
	}
	fmt.Println("Spider complete")

	// Give the passive scanner a chance to complete
	time.Sleep(2000 * time.Millisecond)

	fmt.Println("Active scan : " + target)
	resp, err = client.Ascan().Scan(target, "True", "False", "", "", "", "")
	if err != nil {
		log.Fatal(err)
	}
	// The scan now returns a scan id to support concurrent scanning
	scanid = resp["scan"].(string)
	for {
		time.Sleep(5000 * time.Millisecond)
		resp, _ = client.Ascan().Status(scanid)
		progress, _ := strconv.Atoi(resp["status"].(string))
		fmt.Printf("Active Scan progress : %d\n", progress)
		if progress >= 100 {
			break
		}
	}
	fmt.Println("Active Scan complete")
	fmt.Println("Alerts:")
	alerts, err := client.Core().Alerts(target, "", "", "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", alerts)
	jsonString, err := json.Marshal(alerts)
	if err != nil {
		log.Fatal(err)
	}
	if serve {
		serveResults(jsonString)
	}
}
