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
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/fatih/color"
	"github.com/jhoonb/archivex"
	"github.com/spf13/cobra"
	swagger "github.com/tmaiaroto/aegis/apigateway"
	"github.com/tmaiaroto/aegis/lambda/shim"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

// upCmd is a command that will deploy the app and configuration to AWS Lambda and API Gateway
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy app and API",
	Long:  `Deploys or updates your serverless application and API`,
	Run: func(cmd *cobra.Command, args []string) {
		appPath := ""

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
			err = os.Chmod(appPath, os.FileMode(int(0777)))
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

			cfg.Lambda.SourceZip = compress()
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

		// If no role, create a basic aegis role for executing Lambda functions (this will be very limited role)
		if cfg.Lambda.Role == "" {
			cfg.Lambda.Role = createAegisRole()
			// fmt.Printf("Created a default aegis role for Lambda: %s\n", cfg.Lambda.Role)

			// Have to delay a few seconds to give AWS some time to set up the role.
			// Assigning it to the Lambda too soon could result in an error:
			// InvalidParameterValueException: The role defined for the function cannot be assumed by Lambda.
			// Apparently it needs a few seconds ¯\_(ツ)_/¯
			time.Sleep(4 * time.Second)
		}

		// Create (or update) the function
		lambdaArn := createFunction(zipBytes)

		// Create the API Gateway API with proxy resource.
		// This only needs to be done once as it shouldn't change and additional resources can't be configured.
		// So it will check first for the same name before creating. If a match is found, that API ID will be returned.
		//
		// TODO: Maybe prompt the user to overwrite? Because if the name matches, it will go on to deploy stages on
		// that API...Which may be bad. I wish API names had to be unique. That would be a lot better.
		// Think on what to do here because it could create a bad experience...It's also nice to have one "up" command
		// that also deploys stages and picks up new stages as the config changes. Could always break out deploy stage
		// into a separate command...Again, all comes down to experience and expectations. Warnings might be enough...
		// But a prompt on each "up" command after the first? Maybe too annoying. Could pass an "--ignore" flag or force
		// to solve those annoyances though.
		apiID := importAPI(*lambdaArn)
		// TODO: Allow updates...this isn't quite working yet
		// updateAPI(apiID, *lambdaArn)

		// fmt.Printf("API ID: %s\n", apiID)

		// Ensure the API can access the Lambda
		addAPIPermission(apiID, *lambdaArn)

		// Deploy for each stage (defaults to just one "prod" stage).
		// However, this can be changed over time (cache settings, etc.) and is relatively harmless to re-deploy
		// on each run anyway. Plus, new stages can be added at any time.
		for key := range cfg.API.Stages {
			invokeURL := deployAPI(apiID, cfg.API.Stages[key])
			// fmt.Printf("%s API Invoke URL: %s\n", key, invokeURL)
			fmt.Printf("%v %v %v\n", color.GreenString(key), "API Invoke URL:", color.GreenString(invokeURL))
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

	},
}

// init the `up` command
func init() {
	RootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// aegisAppName is the Go binary built for AWS. The wrapper script refers to this file name. No need to change it.
const aegisAppName = "aegis_app"

// build runs `go build` in the current directory and returns the binary file path to include in the Lambda function zip file.
func build() (string, error) {
	_ = os.Setenv("GOOS", "linux")
	_ = os.Setenv("GOARCH", "amd64")
	path := getExecPath("go")
	pwd, _ := os.Getwd()
	args := []string{path, "build", "-o", aegisAppName}
	cmd := exec.Cmd{
		Path: path,
		Args: args,
	}
	if err := cmd.Run(); err != nil {
		return "", err
	}
	builtApp := filepath.Join(pwd, aegisAppName)

	return builtApp, nil
}

// compress returns the zipped contents of the AWS Lambda function.
func compress() string {
	zipper := new(archivex.ZipFile)
	zipper.Create("aegis_function.zip")

	// Create a header for aegis_app to retain permissions?
	header := &zip.FileHeader{
		Name:         "aegis_app",
		Method:       zip.Store,
		ModifiedTime: uint16(time.Now().UnixNano()),
		ModifiedDate: uint16(time.Now().UnixNano()),
	}
	header.SetMode(os.FileMode(int(0777)))
	_, _ = zipper.Writer.CreateHeader(header)

	// Add the AWS Lambda shim for Go
	// TODO: Allow multiple wrappers
	if err := zipper.Add("index.js", shim.MustAsset(cfg.Lambda.Wrapper)); err != nil {
		return ""
	}

	// Add the compiled Go app
	zipper.AddFile(aegisAppName)
	zipper.Close()

	pwd, _ := os.Getwd()
	builtZip := filepath.Join(pwd, "aegis_function.zip")
	// Set the config
	cfg.Lambda.SourceZip = builtZip

	return builtZip
}

// createFunction will create a Lambda function in AWS
func createFunction(zipBytes []byte) *string {
	svc := lambda.New(session.New(&awsCfg))

	// TODO: Keep versions and allow rollback

	// First check if function already exists
	params := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(cfg.Lambda.FunctionName), // Required
		MaxItems:     aws.Int64(1),
	}
	versionsResp, err := svc.ListVersionsByFunction(params)

	// If there are no previous versions, create the new Lambda function
	if len(versionsResp.Versions) == 0 || err != nil {
		input := &lambda.CreateFunctionInput{
			Code: &lambda.FunctionCode{
				ZipFile: zipBytes,
			},
			Description:  aws.String(cfg.Lambda.Description),
			FunctionName: aws.String(cfg.Lambda.FunctionName),
			Handler:      aws.String(cfg.Lambda.Handler),
			MemorySize:   aws.Int64(cfg.Lambda.MemorySize),
			Publish:      aws.Bool(true),
			Role:         aws.String(cfg.Lambda.Role),
			Runtime:      aws.String(cfg.Lambda.Runtime),
			Timeout:      aws.Int64(int64(cfg.Lambda.Timeout)),
			Environment: &lambda.Environment{
				Variables: cfg.Lambda.EnvironmentVariables,
			},
			KMSKeyArn: aws.String(cfg.Lambda.KMSKeyArn),
			VpcConfig: &lambda.VpcConfig{
				SecurityGroupIds: aws.StringSlice(cfg.Lambda.VPC.SecurityGroups),
				SubnetIds:        aws.StringSlice(cfg.Lambda.VPC.Subnets),
			},
		}
		f, err := svc.CreateFunction(input)
		if err != nil {
			fmt.Println("There was a problem creating the Lambda function.")
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		fmt.Printf("%v %v\n", "Created Lambda function:", color.GreenString(*f.FunctionArn))

		// Create or update alias
		// TODO: This works, but doesn't really help much without roll back support, etc.
		// Might also want another command to adjust the API so it points to a different version and more.
		// Maybe also allowing different stages of the API to use different Lambda versions if that's possible?
		// createOrUpdateAlias(f)

		// return f.FunctionArn
		// Ensure the version number is stripped from the end
		arn := stripLamdaVersionFromArn(*f.FunctionArn)

		fmt.Printf("%v %v %v %v%v\n", "Updated Lambda function:", color.GreenString(arn), "(version ", *f.Version, ")")
		return &arn
	}

	// Otherwise, update the Lambda function
	return updateFunction(zipBytes)
}

// updateFunction will update a Lambda function and its configuration in AWS
func updateFunction(zipBytes []byte) *string {
	svc := lambda.New(session.New(&awsCfg))

	_, err := svc.UpdateFunctionConfiguration(&lambda.UpdateFunctionConfigurationInput{
		Description:  aws.String(cfg.Lambda.Description),
		FunctionName: aws.String(cfg.Lambda.FunctionName),
		Handler:      aws.String(cfg.Lambda.Handler),
		MemorySize:   aws.Int64(cfg.Lambda.MemorySize),
		Role:         aws.String(cfg.Lambda.Role),
		Runtime:      aws.String(cfg.Lambda.Runtime),
		Timeout:      aws.Int64(int64(cfg.Lambda.Timeout)),
		Environment: &lambda.Environment{
			Variables: cfg.Lambda.EnvironmentVariables,
		},
		KMSKeyArn: aws.String(cfg.Lambda.KMSKeyArn),
		VpcConfig: &lambda.VpcConfig{
			SecurityGroupIds: aws.StringSlice(cfg.Lambda.VPC.SecurityGroups),
			SubnetIds:        aws.StringSlice(cfg.Lambda.VPC.Subnets),
		},
	})
	if err != nil {
		fmt.Println("There was a problem updating the Lambda function.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(cfg.Lambda.FunctionName),
		Publish:      aws.Bool(true),
		ZipFile:      zipBytes,
	}
	f, err := svc.UpdateFunctionCode(input)
	if err != nil {
		fmt.Println("There was a problem updating the Lambda function.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	// Create or update alias
	// createOrUpdateAlias(f)

	// Remove the version number at the end.
	arn := stripLamdaVersionFromArn(*f.FunctionArn)

	fmt.Printf("%v %v %v %v%v\n", "Updated Lambda function:", color.GreenString(arn), "(version ", *f.Version, ")")
	return &arn
}

// createOrUpdateAlias will handle the Lambda function alias
func createOrUpdateAlias(f *lambda.FunctionConfiguration) error {
	svc := lambda.New(session.New(&awsCfg))

	_, err := svc.CreateAlias(&lambda.CreateAliasInput{
		FunctionName:    aws.String(cfg.Lambda.FunctionName),
		FunctionVersion: f.Version,
		Name:            aws.String(cfg.Lambda.Alias),
	})
	if err == nil {
		// Successfully created the alias.
		return nil
	}

	if e, ok := err.(awserr.Error); !ok || e.Code() != "ResourceConflictException" {
		return err
	}

	// If here, then the alias was created, but needs to be updated.
	_, err = svc.UpdateAlias(&lambda.UpdateAliasInput{
		FunctionName:    aws.String(cfg.Lambda.FunctionName),
		FunctionVersion: f.Version,
		Name:            aws.String(cfg.Lambda.Alias),
	})
	if err != nil {
		return err
	}

	return nil
}

// deleteFunction will delete a Lambda function in AWS
func deleteFunction(name, version string) {
	svc := lambda.New(session.New(&awsCfg))

	input := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	}
	if len(version) > 0 {
		input.Qualifier = aws.String(version)
	}
	if _, err := svc.DeleteFunction(input); err != nil {
		log.Fatalln(err)
	}
}

// createAegisRole will create a basic role to run Lambda functions if one has not been provided in config
func createAegisRole() string {
	// Default aegis IAM role name: aegis_lambda_function
	// Default aegis IAM policy name: aegis_lambda_logs
	aegisLambdaRoleName := aws.String("aegis_lambda_function")
	aegisLambdaPolicyName := aws.String("aegis_lambda_logs")

	svc := iam.New(session.New(&awsCfg))

	// First see if the role exists
	params := &iam.GetRoleInput{
		RoleName: aegisLambdaRoleName,
	}
	// Don't worry about errors just yet, there'll be more errors below if things aren't configured properly or can't connect.
	resp, err := svc.GetRole(params)
	if err == nil {
		if resp.Role.Arn != nil {
			existingRole := *resp.Role.Arn
			fmt.Printf("%v %v\n", "Using existing execution role for Lambda:", color.GreenString(existingRole))
			return existingRole
		}
	}

	var iamAssumeRolePolicy = `{
	  "Version": "2012-10-17",
	  "Statement": [
	    {
	      "Effect": "Allow",
	      "Principal": {
	        "Service": "lambda.amazonaws.com"
	      },
	      "Action": "sts:AssumeRole"
	    }
	  ]
	}`

	// Create the Lambda execution role
	role, err := svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aegisLambdaRoleName,
		AssumeRolePolicyDocument: aws.String(iamAssumeRolePolicy),
	})
	if err != nil {
		fmt.Println("There was a problem creating a default IAM role for Lambda. Check your configuration.")
		os.Exit(-1)
	}

	var iamLogsPolicy = `{
	  "Version": "2012-10-17",
	  "Statement": [
	    {
	      "Action": [
	        "logs:*"
	      ],
	      "Effect": "Allow",
	      "Resource": "*"
	    }
	  ]
	}`

	// Create the Lambda policy inline
	_, err = svc.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aegisLambdaPolicyName,
		RoleName:       aegisLambdaRoleName,
		PolicyDocument: aws.String(iamLogsPolicy),
	})
	if err != nil {
		fmt.Println("There was a problem creating a default inline IAM policy for Lambda. Check your configuration.")
		fmt.Println(err)
		os.Exit(-1)
	}

	roleArn := *role.Role.Arn
	fmt.Printf("%v %v\n", "Created an execution role for Lambda:", color.GreenString(roleArn))
	return roleArn
}

// importAPI will import an API using Swagger
func importAPI(lambdaArn string) string {
	svc := apigateway.New(session.New(&awsCfg))

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
		Title:     cfg.API.Name,
		LambdaURI: swagger.GetLambdaURI(lambdaArn),
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

// updatAPI will update an API's settings that are not configured in the demployment/stage.
// There is no real need to update the resources or integrations of course, but things like
// the description, name, binary content types, etc. will need to be updated if changed.
func updateAPI(apiID string, lambdaArn string) {
	svc := apigateway.New(session.New(&awsCfg))

	// Build Swagger
	swaggerDefinition, swaggerErr := swagger.NewSwagger(&swagger.SwaggerConfig{
		Title:     cfg.API.Name,
		LambdaURI: swagger.GetLambdaURI(lambdaArn),
	})
	if swaggerErr != nil {
		fmt.Println("There was a problem creating the API.")
		fmt.Println(swaggerErr.Error())
		os.Exit(-1)
	}

	swaggerBytes, err := json.Marshal(swaggerDefinition)
	if err != nil {
		fmt.Println("There was a problem creating the API.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	_, err = svc.PutRestApi(&apigateway.PutRestApiInput{
		Body:           swaggerBytes,
		RestApiId:      aws.String(apiID),
		FailOnWarnings: aws.Bool(false),
		// FailOnWarnings: aws.Bool(true),
		Mode: aws.String("overwrite"),
	})

	if err != nil {
		fmt.Printf("%v %v\n", color.YellowString("Warning: "), "There may have been a problem updating the API.")
		fmt.Println(err.Error())
	}
}

// deployAPI will create a stage and deploy the API
func deployAPI(apiID string, stage deploymentStage) string {
	svc := apigateway.New(session.New(&awsCfg))

	// Must be one of: [58.2, 13.5, 28.4, 237, 0.5, 118, 6.1, 1.6]
	// TODO: Validate user input. Maybe round to nearest value
	if stage.CacheSize == "" {
		stage.CacheSize = apigateway.CacheClusterSize05
	}

	if stage.Cache {
		fmt.Printf("A cache is set for API responses, this will incur additional charges. Cache size is %sGB\n", stage.CacheSize)
	}

	_, err := svc.CreateDeployment(&apigateway.CreateDeploymentInput{
		RestApiId:           aws.String(apiID),      // Required
		StageName:           aws.String(stage.Name), // Required
		CacheClusterEnabled: aws.Bool(stage.Cache),
		CacheClusterSize:    aws.String(stage.CacheSize),
		Description:         aws.String(cfg.API.Description),
		StageDescription:    aws.String(stage.Description),
		Variables:           stage.Variables,
	})
	if err != nil {
		fmt.Println("There was a problem deploying the API.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	// Format the invoke URL
	// https://xxxxx.execute-api.us-east-1.amazonaws.com/prod
	var buffer bytes.Buffer
	buffer.WriteString("https://")
	buffer.WriteString(apiID)
	buffer.WriteString(".execute-api.")
	buffer.WriteString(cfg.AWS.Region)
	buffer.WriteString(".amazonaws.com/")
	buffer.WriteString(stage.Name)
	invokeURL := buffer.String()
	buffer.Reset()

	return invokeURL
}

func addAPIPermission(apiID string, lambdaArn string) {
	// http://stackoverflow.com/questions/39905255/how-can-i-grant-permission-to-api-gateway-to-invoke-lambda-functions-through-clo
	// Glue together this weird SourceArn: arn:aws:execute-api:us-east-1:ACCOUNT_ID:API_ID/*/METHOD/ENDPOINT
	// Not sure if some API call can get it?
	accountID, region := getAccountInfoFromLambdaArn(lambdaArn)

	var buffer bytes.Buffer
	buffer.WriteString("arn:aws:execute-api:")
	buffer.WriteString(region)
	buffer.WriteString(":")
	buffer.WriteString(accountID)
	buffer.WriteString(":")
	buffer.WriteString(apiID)
	// What if ENDPOINT is / ?  ¯\_(ツ)_/¯ will * work?
	buffer.WriteString("/*/ANY/*")
	sourceArn := buffer.String()
	buffer.Reset()

	svc := lambda.New(session.New(&awsCfg))

	// There's no list permissions? So remove first and add.
	// _, err := svc.RemovePermission(&lambda.RemovePermissionInput{
	// 	FunctionName: aws.String("FunctionName"), // Required
	// 	StatementId:  aws.String("StatementId"),  // Required
	// 	Qualifier:    aws.String("Qualifier"),
	// })

	_, err := svc.AddPermission(&lambda.AddPermissionInput{
		Action:       aws.String("lambda:InvokeFunction"),           // Required
		FunctionName: aws.String(cfg.Lambda.FunctionName),           // Required
		Principal:    aws.String("apigateway.amazonaws.com"),        // Required
		StatementId:  aws.String("aegis-api-gateway-invoke-lambda"), // Required
		// EventSourceToken: aws.String("EventSourceToken"),
		// Qualifier:        aws.String("Qualifier"),
		// SourceAccount:    aws.String("SourceOwner"),
		SourceArn: aws.String(sourceArn),
	})
	if err != nil {
		// Ignore "already exists" errors, that's fine. No apparent way to look up permissions before making the add call?
		match, _ := regexp.MatchString("already exists", err.Error())
		if !match {
			fmt.Println("There was a problem setting permissions for API Gateway to invoke the Lambda. Try again or go into AWS console and choose the Lambda function for the integration. It'll be selected already, but re-selecting it again will create this permission behind the scenes. You can not see or set this permission from AWS console manually.")
			fmt.Println(err.Error())
		}
	}
}

// getAccountInfoFromArn will extract the account ID and region from a given ARN
func getAccountInfoFromLambdaArn(lambdaArn string) (string, string) {
	r, _ := regexp.Compile("arn:aws:lambda:(.+):([0-9]+):function")
	matches := r.FindStringSubmatch(lambdaArn)
	accountID := ""
	region := ""
	if len(matches) == 3 {
		region = matches[1]
		accountID = matches[2]
	}

	return accountID, region
}

// stripLamdaVersionFromArn will remove the :123 version number from a given Lambda ARN, which indicates to use the latest version when used in AWS
func stripLamdaVersionFromArn(lambdaArn string) string {
	// arn:aws:lambda:us-east-1:1234567890:function:aegis_example:1
	r, _ := regexp.Compile("arn:aws:lambda:(.+):([0-9]+):function:([A-z0-9\\-\\_]+)($|:[0-9]+)")
	matches := r.FindStringSubmatch(lambdaArn)
	accountID := ""
	region := ""
	functionName := ""
	if len(matches) == 5 {
		region = matches[1]
		accountID = matches[2]
		functionName = matches[3]
		// functionVersion = matches[4]
	}

	var buffer bytes.Buffer
	buffer.WriteString("arn:aws:lambda:")
	buffer.WriteString(region)
	buffer.WriteString(":")
	buffer.WriteString(accountID)
	buffer.WriteString(":function:")
	buffer.WriteString(functionName)
	arn := buffer.String()
	buffer.Reset()

	return arn
}

// getExecPath returns the full path to a passed binary in $PATH.
func getExecPath(name string) string {
	if name == "" {
		log.Println("invalid executable file name")
		os.Exit(-1)
	}
	out, err := exec.Command("which", name).Output()
	if err != nil {
		log.Println("executable file %s not found in $PATH", name)
		os.Exit(-1)
	}
	return string(bytes.TrimSpace(out))
}
