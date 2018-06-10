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
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tmaiaroto/aegis/framework/function"
)

// SrcPath defines the path to the example Go source main.go to be copied from bindata to the current working directory upon init
const SrcPath = "./main.go"

// YmlPath defines the path to the example aegis config file to be copied from bindata to the current working directory upon init
const YmlPath = "./aegis.yaml"

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize app",
	Long:  `Initializes your serverless application and creates a configuration file for you to alter as needed.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := copyConfig(YmlPath)
		if err != nil {
			fmt.Printf("%v %v\n", color.YellowString("Warning:"), err.Error())
		}

		err = copySrc(SrcPath)
		if err != nil {
			fmt.Printf("%v %v\n", color.YellowString("Warning:"), err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// copyConfig will copy a boilerplate aegis.yaml config file to the given path (including file name)
func copyConfig(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return errors.New("An aegis.yaml file already exists in this location. It has been left alone.")
	}
	ioErr := ioutil.WriteFile(filePath, function.MustAsset("example_aegis"), 0644)
	if ioErr != nil {
		//fmt.Printf("%v %v\n", color.RedString("Error:"), ioErr.Error())
		return ioErr
	}
	return nil
}

// copySrc will copy a boilerplate main.go source file to the given path (including file name)
func copySrc(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return errors.New("A main.go file already exists in this location. It has been left alone.")
	}
	ioErr := ioutil.WriteFile(filePath, function.MustAsset("example_main"), 0644)
	if ioErr != nil {
		// fmt.Printf("%v %v\n", color.RedString("Error:"), ioErr.Error())
		return ioErr
	}
	return nil
}
