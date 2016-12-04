// Copyright © 2016 Tom Maiaroto <tom@shift8creative.com>
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
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tmaiaroto/aegis/lambda/function"
	"io/ioutil"
	"os"
)

const SrcPath = "./main.go"
const YmlPath = "./aegis.yaml"

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize app",
	Long:  `Initializes your serverless application and creates a configuration file for you to alter as needed.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(YmlPath); err == nil {
			fmt.Printf("%v %v\n", color.YellowString("Warning: "), "An aegis.yaml file already exists in this location. It has been left alone.")
		} else {
			ymlErr := ioutil.WriteFile(YmlPath, function.MustAsset("example_aegis"), 0644)
			if ymlErr != nil {
				fmt.Printf("%v %v\n", color.RedString("Error: "), ymlErr.Error())
			}
		}

		if _, err := os.Stat(SrcPath); err == nil {
			fmt.Printf("%v %v\n", color.YellowString("Warning: "), "A main.go file already exists in this location. It has been left alone.")
		} else {
			mainErr := ioutil.WriteFile(SrcPath, function.MustAsset("example_main"), 0644)
			if mainErr != nil {
				fmt.Printf("%v %v\n", color.RedString("Error: "), mainErr.Error())
			}
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
