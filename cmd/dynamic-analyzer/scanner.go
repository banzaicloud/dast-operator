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
var openapiURL string

var scannerCmd = &cobra.Command{
	Use:   "scanner",
	Short: "Scanner application using Zap",
	Run: func(cmd *cobra.Command, args []string) {
		scanner()
	},
}

var apiScannerCmd = &cobra.Command{
	Use:   "apiscan",
	Short: "API scanner application using Zap",
	Run: func(cmd *cobra.Command, args []string) {
		apiScanner()
	},
}

func init() {
	scannerCmd.Flags().StringVarP(&zapAddr, "zap-proxy", "p", "http://127.0.0.1:8080", "Zap proxy address")
	scannerCmd.Flags().StringVarP(&target, "target", "t", "http://127.0.0.1:8090/target", "Target address")
	scannerCmd.Flags().StringVarP(&apiKey, "apikey", "a", os.Getenv("ZAPAPIKEY"), "Zap api key")
	scannerCmd.Flags().BoolVarP(&serve, "serve", "s", false, "serve results")
	scannerCmd.Flags().StringVarP(&openapiURL, "openapi", "o", "http://127.0.0.1:8090/swagger.yaml", "Openapi url")
	rootCmd.AddCommand(scannerCmd)
	rootCmd.AddCommand(apiScannerCmd)

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
	summary, err := client.Core().AlertsSummary(target)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("alerts: %v", alerts)
	fmt.Printf("summary: %v", summary)
	jsonString, err := json.Marshal(alerts)
	if err != nil {
		log.Fatal(err)
	}
	if serve {
		serveResults(jsonString)
	}
}

func apiScanner() {
	cfg := &zap.Config{
		Proxy:  zapAddr,
		APIKey: apiKey,
	}
	client, err := zap.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Enable scripts
	client.Script().Load("Alert_on_HTTP_Response_Code_Errors.js", "httpsender", "Oracle Nashorn", "/home/zap/.ZAP_D/scripts/scripts/httpsender/Alert_on_HTTP_Response_Code_Errors.js", "", "")
	client.Script().Enable("Alert_on_HTTP_Response_Code_Errors.js")
	client.Script().Load("Alert_on_Unexpected_Content_Types.js", "httpsender", "Oracle Nashorn", "/home/zap/.ZAP_D/scripts/scripts/httpsender/Alert_on_Unexpected_Content_Types.js", "", "")
	client.Script().Enable("Alert_on_Unexpected_Content_Types.js")

	client.Openapi().ImportUrl(openapiURL, target)
	urls, err := client.Core().Urls(target)
	if err != nil {
		log.Fatal(err)
	}

	if len(urls) == 0 {
		log.Print("Failed to import any URLs")
	}
	resp, err := client.Ascan().Scan(target, "True", "False", "", "", "", "")
	if err != nil {
		log.Fatal(err)
	}
	// The scan now returns a scan id to support concurrent scanning
	scanid := resp["scan"].(string)
	for {
		time.Sleep(5000 * time.Millisecond)
		resp, _ = client.Ascan().Status(scanid)
		progress, _ := strconv.Atoi(resp["status"].(string))
		fmt.Printf("Active API Scan progress : %d\n", progress)
		if progress >= 100 {
			break
		}
	}
	fmt.Println("Active API Scan complete")
	fmt.Println("Alerts:")
	alerts, err := client.Core().Alerts(target, "", "", "")
	if err != nil {
		log.Fatal(err)
	}
	summary, err := client.Core().AlertsSummary(target)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("alerts: %v", alerts)
	fmt.Printf("summary: %v", summary)
}
