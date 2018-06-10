// Copyright © 2016 Tom Maiaroto <tom@SerifAndSemaphore.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmaiaroto/aegis/cmd/deploy"
	// TODO: Make it pretty :)
	// https://github.com/gernest/wow?utm_source=golangweekly&utm_medium=email
)

// updateCmd is a command that will update the AWS Lambda function code, publishing a new version
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update app",
	Long:  `Updates your serverless application, updating the Lambda function only`,
	Run:   Update,
}

// init the `deploy` command
func init() {
	RootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// Update will update an AWS Lambda, intended for quick updates where no configuration changes are needed
func Update(cmd *cobra.Command, args []string) {
	appPath := ""

	// This helps break up many of the functions/steps for deployment
	deployer := deploy.NewDeployer(&cfg, getAWSSession())

	// It is possible to pass a specific zip file from the config instead of building a new one (why would one? who knows, but I liked the pattern of using cfg)
	if cfg.Lambda.SourceZip == "" {
		// Build the Go app in the current directory (for AWS architecture).
		appPath, err := build()
		if err != nil {
			fmt.Println("There was a problem building the Go app for the Lambda function.")
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		// Ensure it's executable.
		// err = os.Chmod(appPath, os.FileMode(int(0777)))
		err = os.Chmod(appPath, os.ModePerm)
		if err != nil {
			fmt.Println("Warning, executable permissions could not be set on Go binary. It may fail to run in AWS.")
			fmt.Println(err.Error())
		}

		// Adjust timestamp?
		// err = os.Chtimes(appPath, time.Now(), time.Now())
		// if err != nil {
		// 	fmt.Println("Warning, executable permissions could not be set on Go binary. It may fail to run in AWS.")
		// 	fmt.Println(err.Error())
		// }

		cfg.Lambda.SourceZip = compress(cfg.App.BuildFileName)
		// If something went wrong, exit
		if cfg.Lambda.SourceZip == "" {
			fmt.Println("There was a problem building the Lambda function zip file.")
			os.Exit(-1)
		}
	}

	// Get the Lambda function zip file's bytes
	var zipBytes []byte
	zipBytes, err := ioutil.ReadFile(cfg.Lambda.SourceZip)
	if err != nil {
		fmt.Println("Could not read from Lambda function zip file.")
		fmt.Println(err)
		os.Exit(-1)
	}

	// Update the function
	err = deployer.UpdateFunctionCode(zipBytes)
	if err != nil {
		fmt.Println("Could not update Lmabda function.")
		fmt.Println(err)
	}

	// Clean up
	if !cfg.App.KeepBuildFiles {
		os.Remove(cfg.Lambda.SourceZip)
		// Remember the Go app may not be built if the source zip file was passed via configuration/CLI flag.
		// However, if it is build then it's for AWS architecture and likely isn't needed by the user. Clean it up.
		// Note: It should be called `aegis_app` to help avoid conflicts.
		if _, err := os.Stat(appPath); err == nil {
			os.Remove(appPath)
		}
	}

}
