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
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/fatih/color"
	"github.com/jhoonb/archivex"
	"github.com/spf13/cobra"
	swagger "github.com/tmaiaroto/aegis/apigateway"
	"github.com/tmaiaroto/aegis/cmd/deploy"
	// TODO: Make it pretty :)
	// https://github.com/gernest/wow?utm_source=golangweekly&utm_medium=email
)

// deployCmd is a command that will deploy the app and configuration to AWS Lambda and API Gateway
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy app and API",
	Long:  `Deploys or updates your serverless application and API`,
	Run:   Deploy,
}

// init the `deploy` command
func init() {
	RootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// aegisAppName is the Go binary built for AWS. The wrapper script refers to this file name. No need to change it.
const aegisAppName = "aegis_app"

// Deploy will build and deploy to AWS Lambda and API Gateway
func Deploy(cmd *cobra.Command, args []string) {
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

	// If no role, create an aegis role for executing Lambda functions.
	// This will actually be rather permissive. Use a custom role to be more restrictive.
	// The aegis framework needs the ability to invoke other Lambdas, work with XRay, S3, and more.
	// So it's going to be a few managed policies that make sense. Use a custom role if needed.
	// When roles are passed to use, they are not modified.
	if cfg.Lambda.Role == "" {
		cfg.Lambda.Role = createOrUpdateAegisRole()
		// fmt.Printf("Created a default aegis role for Lambda: %s\n", cfg.Lambda.Role)

		// Have to delay a few seconds to give AWS some time to set up the role.
		// Assigning it to the Lambda too soon could result in an error:
		// InvalidParameterValueException: The role defined for the function cannot be assumed by Lambda.
		// Apparently it needs a few seconds ¯\_(ツ)_/¯
		time.Sleep(4 * time.Second)
	}

	// Create (or update) the function
	lambdaArn := deployer.CreateFunction(zipBytes)

	// Create the API Gateway API with proxy resource.
	// This only needs to be done once as it shouldn't change and additional resources can't be configured.
	// So it will check first for the same name before creating. If a match is found, that API ID will be returned.
	//
	// TODO: Maybe prompt the user to overwrite? Because if the name matches, it will go on to deploy stages on
	// that API...Which may be bad. I wish API names had to be unique. That would be a lot better.
	// Think on what to do here because it could create a bad experience...It's also nice to have one "deploy" command
	// that also deploys stages and picks up new stages as the config changes. Could always break out deploy stage
	// into a separate command...Again, all comes down to experience and expectations. Warnings might be enough...
	// But a prompt on each "deploy" command after the first? Maybe too annoying. Could pass an "--ignore" flag or force
	// to solve those annoyances though.
	apiID := deployer.ImportAPI(*lambdaArn)
	// TODO: Allow updates...this isn't quite working yet
	// The DeployAPI() function will take care of some updates as well (things like stage variables, etc.).
	// deployer.UpdateAPI(apiID, *lambdaArn)

	// fmt.Printf("API ID: %s\n", apiID)

	// Ensure the API can access the Lambda
	deployer.AddAPIPermission(apiID, *lambdaArn)

	// Ensure the API has it's binary media types set (Swagger import apparently does not set them)
	deployer.AddBinaryMediaTypes(apiID)

	// Deploy for each stage (defaults to just one "prod" stage).
	// However, this can be changed over time (cache settings, stage variables, etc.) and is relatively harmless to re-deploy
	// on each run anyway. Plus, new stages can be added at any time.
	for key := range cfg.API.Stages {
		invokeURL := deployer.DeployAPI(apiID, cfg.API.Stages[key])
		// fmt.Printf("%s API Invoke URL: %s\n", key, invokeURL)
		fmt.Printf("%v %v %v\n", color.GreenString(key), "API URL:", color.GreenString(invokeURL))
	}

	// Tasks (CloudWatch event rules to trigger Lambda)
	fmt.Printf("\n")
	deployer.AddTasks()

	// Bucket notifications (to trigger Lambda)
	fmt.Printf("\n")
	deployer.AddS3BucketNotifications()

	// SES Recipient Rules (to trigger Lambda)
	if cfg.SESRules != nil && len(cfg.SESRules) > 0 {
		fmt.Printf("\n")
		deployer.AddSESPermission(lambdaArn)
		deployer.AddSESRules()
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

// build runs `go build` in the current directory and returns the binary file path to include in the Lambda function zip file.
func build() (string, error) {
	_ = os.Setenv("GOOS", "linux")
	_ = os.Setenv("GOARCH", "amd64")
	path := getExecPath("go")
	pwd, _ := os.Getwd()

	// Try to build a smaller binary.
	// This won't work on Windows. Though Windows remains untested in general, let's try this and fall back.
	cmd := exec.Command("sh", "-c", path+` build -ldflags="-w -s" -o `+aegisAppName)
	if err := cmd.Run(); err != nil {
		// If it failed, just build without all the fancy flags. The binary size will be a little larger though.
		// This should work on Windows. Right? TODO: Test. Better yet, figure out how to build Cmd with flags.
		// Spent over an hour trying every method of escaping known to man. Why???
		args := []string{path, "build", "-o", aegisAppName}
		cmd := exec.Cmd{
			Path: path,
			Args: args,
		}
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}
	builtApp := filepath.Join(pwd, aegisAppName)

	return builtApp, nil
}

// compress zips the AWS Lambda function files and returns the zip file path.
func compress(fileName string) string {
	zipper := new(archivex.ZipFile)
	zipper.Create(fileName)

	// Setting permissions on file to os.ModePerm or 0777 doesn't seem to keep the proper permissions.
	// So use the file's own?
	// aegisAppFileInfo, _ := os.Stat(aegisAppName)

	// Create a header for aegis_app to retain permissions?
	header := &zip.FileHeader{
		Name:         "aegis_app",
		Method:       zip.Store,
		ModifiedTime: uint16(time.Now().UnixNano()),
		ModifiedDate: uint16(time.Now().UnixNano()),
	}
	// os.ModePerm = -rwxrwxrwx
	header.SetMode(os.ModePerm)
	// log.Println("aegis_app file mode:", aegisAppFileInfo.Mode())
	// header.SetMode(aegisAppFileInfo.Mode())
	zipWriter, _ := zipper.Writer.CreateHeader(header)
	// log.Println("zip header", header)

	content, err := ioutil.ReadFile(aegisAppName)
	if err == nil {
		zipWriter.Write(content)
	}

	// Add the compiled Go app
	// zipper.AddFile(aegisAppName) <-- maybe this is writing w/o the header... have to use the writer returned by CreateHeader()??
	zipper.Close()

	pwd, _ := os.Getwd()
	builtZip := filepath.Join(pwd, fileName)
	// Set the config
	cfg.Lambda.SourceZip = builtZip

	return builtZip
}

// setCredentials will try to set AWS credentials from a variety of methods
func setCredentials() *credentials.Credentials {
	// First, try the credentials file set by AWS CLI tool.
	// Note the empty string instructs to look under default file path (different based on OS).
	// This file can have multiple profiles and a default profile will be used unless otherwise configured.
	// See: https://godoc.org/github.com/aws/aws-sdk-go/aws/credentials#SharedCredentialsProvider
	creds := credentials.NewSharedCredentials("", cfg.AWS.Profile)

	// Second, use environment variables if set. The following are checked:
	// Access Key ID: AWS_ACCESS_KEY_ID or AWS_ACCESS_KEY
	// Secret Access Key: AWS_SECRET_ACCESS_KEY or AWS_SECRET_KEY
	envCreds := credentials.NewEnvCredentials()
	setCreds, _ := envCreds.Get()
	// error apparently does not return if environment variables weren't set
	// so check what was set and look for empty strings, don't want to set empty creds
	if setCreds.AccessKeyID != "" && setCreds.SecretAccessKey != "" {
		creds = envCreds
	}

	// Last, if credentials were passed via CLI, always prefer those
	if cfg.AWS.AccessKeyID != "" && cfg.AWS.SecretAccessKey != "" {
		creds = credentials.NewStaticCredentials(cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey, "")
	}

	return creds
}

// createOrUpdateAegisRole will manage a basic role to run Lambda functions if one has not been provided in config
func createOrUpdateAegisRole() string {
	// Default aegis IAM role name: aegis_lambda_role
	aegisLambdaRoleName := aws.String("aegis_lambda_role")
	roleArn := ""

	svc := iam.New(getAWSSession())

	// First see if the role exists
	params := &iam.GetRoleInput{
		RoleName: aegisLambdaRoleName,
	}
	// Don't worry about errors just yet, there'll be more errors below if things aren't configured properly or can't connect.
	resp, err := svc.GetRole(params)
	if err == nil {
		if resp.Role.Arn != nil {
			roleArn = *resp.Role.Arn
			fmt.Printf("%v %v\n", "Using existing execution role for Lambda:", color.GreenString(roleArn))
		}
	}

	// Create the Lambda execution role, if necessary
	if roleArn == "" {
		var iamAssumeRolePolicy = `{
			"Version": "2012-10-17",
			"Statement": [
			  {
				"Effect": "Allow",
				"Principal": {
				  "Service": "lambda.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			  },
			  {
				"Effect": "Allow",
				"Principal": {
				  "Service": "events.amazonaws.com"
				},
				  "Action": "sts:AssumeRole"
			  },
			  {
				"Effect": "Allow",
				"Principal": {
				  "Service": "cognito-identity.amazonaws.com"
				},
				  "Action": "sts:AssumeRole"
			  },
			  {
				"Effect": "Allow",
				"Principal": {
				  "Service": "xray.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			  }
			]
		  }`

		role, err := svc.CreateRole(&iam.CreateRoleInput{
			RoleName:                 aegisLambdaRoleName,
			AssumeRolePolicyDocument: aws.String(iamAssumeRolePolicy),
		})
		if err != nil {
			fmt.Println("There was a problem creating a default IAM role for Lambda. Check your configuration.")
			os.Exit(-1)
		}
		roleArn := *role.Role.Arn
		fmt.Printf("%v %v\n", "Created a new execution role for Lambda:", color.GreenString(roleArn))
	}

	// Attach managed policies.
	// First, AWSLambdaFullAccess
	_, err = svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AWSLambdaFullAccess"),
		RoleName:  aegisLambdaRoleName,
	})
	if err != nil {
		fmt.Println("There was a problem attaching AWSLambdaFullAccess managed policy to the IAM role for Lambda.")
		fmt.Println(err)
	}

	// Then AmazonCognitoReadOnly
	_, err = svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonCognitoReadOnly"),
		RoleName:  aegisLambdaRoleName,
	})
	if err != nil {
		fmt.Println("There was a problem attaching AmazonCognitoReadOnly managed policy to the IAM role for Lambda.")
		fmt.Println(err)
	}

	// Then AWSXrayFullAccess
	_, err = svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AWSXrayFullAccess"),
		RoleName:  aegisLambdaRoleName,
	})
	if err != nil {
		fmt.Println("There was a problem attaching AWSXrayFullAccess managed policy to the IAM role for Lambda.")
		fmt.Println(err)
	}

	return roleArn
}

// importAPI will import an API using Swagger
func importAPI(lambdaArn string) string {
	svc := apigateway.New(getAWSSession())

	// First check to see if there's already an API by the same name
	// (only pulls up to 100 APIs, so this isn't a great long term solution)
	apisResp, err := svc.GetRestApis(&apigateway.GetRestApisInput{
		Limit: aws.Int64(100),
	})
	if err != nil {
		fmt.Println("There was a problem creating the API.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	for key := range apisResp.Items {
		if *apisResp.Items[key].Name == cfg.API.Name {
			// TODO: Prompt user to continue and add a new API anyway. Or remove/overwrite/ignore?
			// Inspect the same named APIs and see if there's a {proxy+} path?
			// It's possible to have multiple APIs with the same name. I hate to break this into
			// multiple commands/steps, it's nice just running `up` and nothing else...But it's not
			// perfect because the user doesn't set the unique identifier for the API.
			fmt.Println("API already exists.")
			return *apisResp.Items[key].Id
		}
	}

	// Build Swagger
	swaggerDefinition, swaggerErr := swagger.NewSwagger(&swagger.SwaggerConfig{
		Title:             cfg.API.Name,
		LambdaURI:         swagger.GetLambdaURI(lambdaArn),
		ResourceTimeoutMs: cfg.API.ResourceTimeoutMs,
		// BinaryMediaTypes: cfg.API.BinaryMediaTypes,
	})
	if swaggerErr != nil {
		fmt.Println(swaggerErr.Error())
		os.Exit(-1)
	}

	swaggerBytes, err := json.Marshal(swaggerDefinition)
	if err != nil {
		fmt.Println("There was a problem creating the API.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	// Import from Swagger
	resp, err := svc.ImportRestApi(&apigateway.ImportRestApiInput{
		Body:           swaggerBytes, // Required
		FailOnWarnings: aws.Bool(true),
	})
	if err != nil {
		fmt.Println("There was a problem creating the API.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	return *resp.Id
}

// getExecPath returns the full path to a passed binary in $PATH.
func getExecPath(name string) string {
	if name == "" {
		log.Println("invalid executable file name")
		os.Exit(-1)
	}
	out, err := exec.Command("which", name).Output()
	if err != nil {
		log.Printf("executable file %s not found in $PATH", name)
		os.Exit(-1)
	}
	return string(bytes.TrimSpace(out))
}
