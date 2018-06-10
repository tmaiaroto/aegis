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
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/lambda"
	// "github.com/fatih/color"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmaiaroto/aegis/cmd/config"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "aegis",
	Short: "Deploy RESTful serverless APIs ",
	Long:  `A tool to deploy a RESTful serverless API using AWS Lambda and API Gateway with a Lambda Proxy and an API Gateway ANY request.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// cfg holds the Aegis configuration for the Lambda function, API Gateway settings, etc.
var cfg config.DeploymentConfig

// awsCfg holds the AWS configuration and credentials
var awsCfg aws.Config

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "aegis", "config file (default is aegis.yaml)")

	// AWS options & credentials
	RootCmd.PersistentFlags().StringVarP(&cfg.AWS.Region, "region", "r", "us-east-1", "AWS Region to use")
	RootCmd.PersistentFlags().StringVarP(&cfg.AWS.AccessKeyID, "keyId", "k", "", "AWS Access Key ID")
	RootCmd.PersistentFlags().StringVarP(&cfg.AWS.SecretAccessKey, "secretKey", "s", "", "AWS Secret Access Key")
	RootCmd.PersistentFlags().StringVarP(&cfg.AWS.Profile, "profile", "p", "default", "AWS Credentials Profile to use")
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	viper.SetConfigName("aegis") // name of config file (without extension)
	viper.AddConfigPath(".")
	// viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv() // read in environment variables that match

	// Default config values
	viper.SetDefault("aws.region", "us-east-1")
	// viper.SetDefault("aws.profile", "default") // set by defaults on flags
	// Default Lambda config values
	viper.SetDefault("lambda.runtime", lambda.RuntimeGo1X)
	// Aegis will build a Go binary named aegis_app ... This shouldn't need to change.
	viper.SetDefault("lambda.handler", "aegis_app")
	viper.SetDefault("lambda.alias", "current")
	// In megabytes
	viper.SetDefault("lambda.memorySize", int64(128))
	// In seconds
	viper.SetDefault("lambda.timeout", int64(3))
	// Don't even default it. Let it be, it'll use the current account max.
	// viper.SetDefault("lambda.maxConcurrentExecutions", int64());
	// No suitable default for this
	// viper.SetDefault("lambda.role", "arn:aws:iam::account-id:role/lambda_basic_execution")
	// Set a default function name
	fName := "aegis_function"
	// Use the current directory name by default
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		// Prepend aegis_
		if dir != "/" && dir != "" {
			dirParts := strings.Split(dir, "/")
			dir = dirParts[len(dirParts)-1]
			var buffer bytes.Buffer
			buffer.WriteString("aegis_")
			buffer.WriteString(dir)
			fName = buffer.String()
			buffer.Reset()
		}
	}
	viper.SetDefault("lambda.functionName", fName)
	// Enable XRay by default, see: https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#TracingConfig
	viper.SetDefault("lambda.traceMode", "Active")
	// Default API Gateway config values
	viper.SetDefault("api.name", "Aegis API")
	viper.SetDefault("api.description", "")
	viper.SetDefault("api.cache", false)
	viper.SetDefault("api.binaryMediaTypes", "*/*")
	// For valid values, see: https://godoc.org/github.com/aws/aws-sdk-go/service/apigateway#pkg-constants
	viper.SetDefault("api.cacheSize", apigateway.CacheClusterSize05)
	// All resources ({proxy+} and /) get this timeout in ms (AWS default is 29000 too)
	viper.SetDefault("api.resourceTimeoutMs", 29000)

	// Default API stage (does not use caching, that comes with an additional cost)
	viper.SetDefault("api.stages", map[string]config.DeploymentStage{
		"prod": config.DeploymentStage{
			Name:        "prod",
			Description: "production stage",
			Cache:       false, // no cache by default
			// CacheSize:   apigateway.CacheClusterSize05, // never needed because above is false and empty value caught in deployment
		},
	})

	// By default do not keep the build files (clean up)
	viper.SetDefault("app.keepBuildFiles", false)
	// Just in case the temporary zip file that gets built creates a conflict, it can be adjusted. However, this is the default.
	viper.SetDefault("app.buildFileName", "aegis_function.zip")

	// If a config file is found, read it in.
	err = viper.ReadInConfig()
	if err == nil {
		// TODO verbose?
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		// Not strictly required given defaults? This prevents other commands such as `init` from working.
		// fmt.Printf("%v %v\n", color.YellowString("Warning:"), "Could not find aegis config file.")
		// os.Exit(-1)
	}

	_ = viper.Unmarshal(&cfg)
	// if err != nil {
	// 	fmt.Println("Could parse aegis config file.")
	// 	os.Exit(-1)
	// }

	// Initialize AWS config
	awsCfg = aws.Config{
		Region: aws.String(cfg.AWS.Region),
	}
}

// SetAwsCfg will set awsCfg values using an aws.Config struct
func SetAwsCfg(config aws.Config) {
	awsCfg = config
}

// getAWSSession will return a session based on options passed to aegis
func getAWSSession() *session.Session {
	// get new credentials if not set
	if awsCfg.Credentials == nil {
		awsCfg.Credentials = setCredentials()
	}

	// session options
	opts := session.Options{
		Config:  awsCfg,
		Profile: cfg.AWS.Profile,
	}

	// Note: New() has been deprecated from aws-sdk-go
	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		fmt.Println("There was a problem creating a session with AWS. Make sure you have credentials configured.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	return sess
}
