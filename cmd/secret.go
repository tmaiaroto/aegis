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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// secretCmd represents the secret command parent
var secretCmd = &cobra.Command{Use: "secret"}

// secretStoreCmd represents the subcommand to store secrets
var secretStoreCmd = &cobra.Command{
	Use:   "store [name] [key] [value]",
	Short: "Store new secret",
	Long:  `Stores a new secret into AWS Secret Manager.`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		// SecretBinary is a pretty interesting field under CreateSecretInput
		// Only available to AWS CLI or SDK.
		// Tags is another interesting one.

		kv := map[string]interface{}{args[1]: args[2]}
		kvB, err := json.Marshal(kv)
		if err == nil {
			// TODO: aws.NewConfig().WithRegion("us-west-2")
			// Think about flags for region... As a whole here.
			// It's in the aegis.yaml ...
			// So I don't know if when using secrets they need to be from same region...
			// Or if we can make <region.secretName.keyName> values... Which might be kinda neat.
			// Certainly can start without that and always add it in a backwards compatible way.
			svc := secretsmanager.New(getAWSSession())
			_, err := svc.CreateSecret(&secretsmanager.CreateSecretInput{
				Name:         aws.String(args[0]),
				SecretString: aws.String(string(kvB)),
			})

			if err != nil {
				if strings.Contains(err.Error(), "ResourceExistsException") {
					prompt := promptui.Prompt{
						Label:     args[0] + " already exists, append key/value to this secret?",
						IsConfirm: true,
					}

					result, err := prompt.Run()
					if err != nil {
						return
					}
					if result == "y" {
						existingValue, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
							// Can be either ARN or friendly name (fortunately)
							SecretId: aws.String(args[0]),
						})
						if err == nil {
							var kvE map[string]interface{}
							if err := json.Unmarshal([]byte(aws.StringValue(existingValue.SecretString)), &kvE); err != nil {
								fmt.Println("There was a problem unmarshaling the secret value. Are you trying to append to a binary secret value?", err)
							}
							// Edit existing value
							for k, v := range kv {
								kvE[k] = v
							}
							// Marshal back
							kvB, err = json.Marshal(kvE)
							if err != nil {
								fmt.Println("There was a problem storing the secret.", err)
								return
							}

							_, err := svc.UpdateSecret(&secretsmanager.UpdateSecretInput{
								SecretId:     aws.String(args[0]),
								SecretString: aws.String(string(kvB)),
							})
							if err != nil {
								fmt.Println("There was a problem updating the secret.", err)
							} else {
								fmt.Println("Successfully appended key/value to secret.")
							}
						} else {
							fmt.Println("There was a problem looking up the existing secret in order to append to it.", err)
						}
					}

				} else {
					fmt.Println("There was a problem storing the secret.", err)
				}
			} else {
				fmt.Println("Successfully stored secret.")
			}
		}
	},
}

// secretReadCmd represents the subcommand to read secrets
var secretReadCmd = &cobra.Command{
	Use:   "read [name] [key]",
	Short: "Reads a secret, partly hidden for security",
	Long:  `Reads one or more of a secret's key values from AWS Secret Manager, hiding parts for security.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If getting all key values from a secret
		if len(args) == 1 {
			values := getSecretValue(args[0])
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Key", "Value"})
			// Don't wrap text, that's bad for copy/paste and reading important values
			table.SetAutoWrapText(false)

			// Sort by key name
			var keys []string
			for k := range values {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				table.Append([]string{k, hidePartOfString(fmt.Sprintf("%v", values[k]), color.HiBlackString("*"))})
			}

			table.Render()
		}

		// If getting just a single key value from a secret
		if len(args) > 1 {
			val := getSecretKeyValue(args[0], args[1])
			if val == "" {
				return
			}

			// Hide half the value if the string is long enough
			// hiddenVal := hidePartOfString(val)
			hiddenVal := hidePartOfString(val, color.HiBlackString("*"))
			valRune := []rune(hiddenVal)
			charLen := len(valRune)
			if (charLen / 2) > 0 {
				fmt.Printf("%v: %v", args[1], hiddenVal)
			} else {
				prompt := promptui.Prompt{
					Label:     "This value is not long enough to hide. Show it anyway?",
					IsConfirm: true,
				}
				result, err := prompt.Run()
				if err != nil {
					return
				}
				if result == "y" {
					fmt.Printf("%v: %v", args[1], val)
				}
			}
		}

	},
}

// secretReadFullCmd represents the subcommand to read secrets
var secretReadFullCmd = &cobra.Command{
	Use:   "full [name] [key]",
	Short: "Reads a secret completely",
	Long:  `Reads one or more of a secret's key values from AWS Secret Manager, showing the complete values.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If getting all key values from a secret
		if len(args) == 1 {
			values := getSecretValue(args[0])
			// fmt.Printf("%v", values)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Key", "Value"})
			// Don't wrap text, that's bad for copy/paste and reading important values
			table.SetAutoWrapText(false)

			// Sort by key name
			var keys []string
			for k := range values {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				table.Append([]string{k, fmt.Sprintf("%v", values[k])})
			}

			table.Render()
		}

		// If getting just a single key value from a secret
		if len(args) > 1 {
			val := getSecretKeyValue(args[0], args[1])
			if val != "" {
				fmt.Printf("%v: %v", args[1], val)
			}
		}
	},
}

// getSecretValue returns an entire secret, returning a map, prints error messages
func getSecretValue(secretName string) map[string]interface{} {
	var kvE map[string]interface{}
	svc := secretsmanager.New(getAWSSession())
	existingValue, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		// Can be either ARN or friendly name (fortunately)
		SecretId: aws.String(secretName),
	})
	if err == nil {
		if err := json.Unmarshal([]byte(aws.StringValue(existingValue.SecretString)), &kvE); err != nil {
			fmt.Println("There was a problem unmarshaling the secret value. Perhaps this is a binary secret value?", err)
		}
	} else {
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			fmt.Println("Secret not found.")
		} else {
			fmt.Println("Could not read secret.", err)
		}
	}
	return kvE
}

// getSecretKeyValue returns a key's value or empty string, prints error messages
func getSecretKeyValue(secretName string, keyName string) string {
	keyValue := ""
	kvE := getSecretValue(secretName)
	if val, ok := kvE[keyName]; ok {
		keyValue = val.(string)
	} else {
		fmt.Println("That key does not exist for " + secretName + ".")
	}
	return keyValue
}

// hidePartOfString will hide half of a string's value, replacing characters with the replacement (asterisk by default).
// Note if value is only one character, it will be returned as is. It doens't make sense to hide.
func hidePartOfString(val string, replacement ...string) string {
	// Hide half the value if the string is long enough
	valRune := []rune(val)
	charLen := len(valRune)
	if (charLen / 2) > 0 {
		valRune = append(valRune[:0], valRune[(charLen/2):]...)
		protectedValRune := []rune("")
		for i := 1; i <= (charLen / 2); i++ {
			// If no replacement value was provided
			if len(replacement) == 0 {
				protectedValRune = append(protectedValRune, '*')
			} else {
				// A replacement can be given as a string, but we need to convert to rune.
				// ie. color.Cyantring("*")
				rep := []rune(replacement[0])
				for _, v := range rep {
					protectedValRune = append(protectedValRune, v)
				}
			}
		}
		protectedValRune = append(protectedValRune, valRune...)
		val = string(protectedValRune)
		// fmt.Printf("%v: %v", args[1], val)
	}
	return val
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	secretStoreCmd.Flags().StringP("description", "d", "", "Secret description (optional)")
	secretStoreCmd.Flags().StringP("kmsARN", "i", "", "KMS Key ID ARN (optional)")

	secretCmd.AddCommand(secretStoreCmd, secretReadCmd, secretReadFullCmd)
	RootCmd.AddCommand(secretCmd)
}
