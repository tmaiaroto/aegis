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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/spf13/cobra"
)

// appSyncCmd represents the secret command parent
var appSyncCmd = &cobra.Command{Use: "appsync"}

// appSyncImportCmd represents the subcommand to store secrets
var appSyncImportCmd = &cobra.Command{
	Use:   "import [id]",
	Short: "Import schema",
	Long:  `Given an ID, imports an AppSync GraphQL schema into your Aegis project.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		svc := appsync.New(getAWSSession())
		graphQLAPI, err := svc.GetGraphqlApi(&appsync.GetGraphqlApiInput{
			ApiId: aws.String(args[0]),
		})
		if err == nil {
			fmt.Println(graphQLAPI)

			// types, err := svc.ListTypes(&appsync.ListTypesInput{
			// 	ApiId:  aws.String(args[0]),
			// 	Format: aws.String("SDL"),
			// })
			// if err == nil {
			// 	fmt.Println(types)
			// } else {
			// 	fmt.Println("Could not get types.", err)
			// }

			schema, _ := svc.GetIntrospectionSchema(&appsync.GetIntrospectionSchemaInput{
				ApiId:  aws.String(args[0]),
				Format: aws.String("SDL"),
			})
			fmt.Println(string(schema.Schema))
		} else {
			fmt.Println("Something went wrong, did you specify a valid AppSync ID?", err)
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	appSyncCmd.AddCommand(appSyncImportCmd)
	RootCmd.AddCommand(appSyncCmd)
}
