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

// Package deploy is purely for organization, the deploy.go command file was getting absurdly long
package deploy

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/fatih/color"
	"github.com/tmaiaroto/aegis/cmd/config"
	"github.com/tmaiaroto/aegis/cmd/util"
)

// Deployer will hold a DeploymentConfig to use with its various functions for deployment
type Deployer struct {
	Cfg        *config.DeploymentConfig
	AWSSession *session.Session
	LambdaArn  *string
	TasksPath  string
}

// NewDeployer takes a cfg argument to set the config needed for its various functions
func NewDeployer(cfg *config.DeploymentConfig, session *session.Session) *Deployer {
	d := Deployer{
		Cfg:        cfg,
		AWSSession: session,
		// TasksPath defines the path to look for CloudWatch Event Rules ("tasks") defined in JSON files
		// Not currently set via Cfg, but can be changed in this interface
		TasksPath: "./tasks",
	}
	return &d
}

// CreateFunction will create a Lambda function in AWS and return its ARN
func (d *Deployer) CreateFunction(zipBytes []byte) *string {
	svc := lambda.New(d.AWSSession)
	// TODO: Keep versions and allow rollback

	// First check if function already exists
	params := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(d.Cfg.Lambda.FunctionName), // Required
		MaxItems:     aws.Int64(1),
	}
	versionsResp, err := svc.ListVersionsByFunction(params)

	// If there are no previous versions, create the new Lambda function
	if len(versionsResp.Versions) == 0 || err != nil {
		input := &lambda.CreateFunctionInput{
			Code: &lambda.FunctionCode{
				ZipFile: zipBytes,
			},
			Description:  aws.String(d.Cfg.Lambda.Description),
			FunctionName: aws.String(d.Cfg.Lambda.FunctionName),
			Handler:      aws.String(d.Cfg.Lambda.Handler),
			MemorySize:   aws.Int64(d.Cfg.Lambda.MemorySize),
			Publish:      aws.Bool(true),
			Role:         aws.String(d.Cfg.Lambda.Role),
			Runtime:      aws.String(d.Cfg.Lambda.Runtime),
			Timeout:      aws.Int64(int64(d.Cfg.Lambda.Timeout)),
			Environment: &lambda.Environment{
				// Variables: d.Cfg.Lambda.EnvironmentVariables,
				Variables: d.LookupSecretsForLambdaEnvVars(d.Cfg.Lambda.EnvironmentVariables),
			},
			KMSKeyArn: aws.String(d.Cfg.Lambda.KMSKeyArn),
			VpcConfig: &lambda.VpcConfig{
				SecurityGroupIds: aws.StringSlice(d.Cfg.Lambda.VPC.SecurityGroups),
				SubnetIds:        aws.StringSlice(d.Cfg.Lambda.VPC.Subnets),
			},
			TracingConfig: &lambda.TracingConfig{
				Mode: aws.String(d.Cfg.Lambda.TraceMode),
			},
		}
		f, err := svc.CreateFunction(input)
		if err != nil {
			fmt.Println("There was a problem creating the Lambda function.")
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		fmt.Printf("%v %v\n\n", "Created Lambda function:", color.GreenString(*f.FunctionArn))

		// Create or update alias
		// TODO: This works, but doesn't really help much without roll back support, etc.
		// Might also want another command to adjust the API so it points to a different version and more.
		// Maybe also allowing different stages of the API to use different Lambda versions if that's possible?
		// createOrUpdateAlias(f)

		// return f.FunctionArn
		// Ensure the version number is stripped from the end
		arn := util.StripLamdaVersionFromArn(*f.FunctionArn)

		// Set concurrency limit, if configured
		d.updateFunctionMaxConcurrency(svc)

		// Believe this is in error. Or rather I think it was related to creating/updating an alias.
		// fmt.Printf("%v %v %v %v%v\n\n", "Updated Lambda function:", color.GreenString(arn), "(version ", *f.Version, ")")

		d.LambdaArn = &arn
		return &arn
	}

	// Otherwise, update the Lambda function
	updatedArn := d.updateFunction(zipBytes)
	d.LambdaArn = updatedArn
	return updatedArn
}

// updateFunction will update a Lambda function and its configuration in AWS and return its ARN
func (d *Deployer) updateFunction(zipBytes []byte) *string {
	svc := lambda.New(d.AWSSession)

	_, err := svc.UpdateFunctionConfiguration(&lambda.UpdateFunctionConfigurationInput{
		Description:  aws.String(d.Cfg.Lambda.Description),
		FunctionName: aws.String(d.Cfg.Lambda.FunctionName),
		Handler:      aws.String(d.Cfg.Lambda.Handler),
		MemorySize:   aws.Int64(d.Cfg.Lambda.MemorySize),
		Role:         aws.String(d.Cfg.Lambda.Role),
		Runtime:      aws.String(d.Cfg.Lambda.Runtime),
		Timeout:      aws.Int64(int64(d.Cfg.Lambda.Timeout)),
		Environment: &lambda.Environment{
			// Variables: d.Cfg.Lambda.EnvironmentVariables,
			Variables: d.LookupSecretsForLambdaEnvVars(d.Cfg.Lambda.EnvironmentVariables),
		},
		KMSKeyArn: aws.String(d.Cfg.Lambda.KMSKeyArn),
		VpcConfig: &lambda.VpcConfig{
			SecurityGroupIds: aws.StringSlice(d.Cfg.Lambda.VPC.SecurityGroups),
			SubnetIds:        aws.StringSlice(d.Cfg.Lambda.VPC.Subnets),
		},
		TracingConfig: &lambda.TracingConfig{
			Mode: aws.String(d.Cfg.Lambda.TraceMode),
		},
	})
	if err != nil {
		fmt.Println("There was a problem updating the Lambda function.")
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(d.Cfg.Lambda.FunctionName),
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
	arn := util.StripLamdaVersionFromArn(*f.FunctionArn)

	// Set concurrency limit, if configured
	d.updateFunctionMaxConcurrency(svc)

	fmt.Printf("%v %v %v %v%v\n\n", "Updated Lambda function:", color.GreenString(arn), "(version", *f.Version, ")")
	return &arn
}

// UpdateFunctionCode updates the Lambda function code and publishes a new version - no configuration changes
func (d *Deployer) UpdateFunctionCode(zipBytes []byte) error {
	// https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#Lambda.UpdateFunctionCode
	svc := lambda.New(d.AWSSession)

	f, err := svc.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(d.Cfg.Lambda.FunctionName),
		Publish:      aws.Bool(true),
		ZipFile:      zipBytes,
	})

	if err == nil {
		fmt.Printf("Updated Lambda function: %v %v %v%v\n\n", color.GreenString(d.Cfg.Lambda.FunctionName), "(version", *f.Version, ")")
	}
	return err
}

// createOrUpdateAlias will handle the Lambda function alias
// TODO: Unused. Think about implementing this
func (d *Deployer) createOrUpdateAlias(f *lambda.FunctionConfiguration) error {
	svc := lambda.New(d.AWSSession)

	_, err := svc.CreateAlias(&lambda.CreateAliasInput{
		FunctionName:    aws.String(d.Cfg.Lambda.FunctionName),
		FunctionVersion: f.Version,
		Name:            aws.String(d.Cfg.Lambda.Alias),
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
		FunctionName:    aws.String(d.Cfg.Lambda.FunctionName),
		FunctionVersion: f.Version,
		Name:            aws.String(d.Cfg.Lambda.Alias),
	})
	if err != nil {
		return err
	}

	return nil
}

// deleteFunction will delete a Lambda function in AWS
// TODO: Unused. Should the CLI be destructive?
func (d *Deployer) deleteFunction(name, version string) {
	svc := lambda.New(d.AWSSession)

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

// updateFunctionMaxConcurrency will adjust the concurrency, if configured
func (d *Deployer) updateFunctionMaxConcurrency(svc *lambda.Lambda) {
	// is function name the arn??? Or is it aws.String(cfg.Lambda.FunctionName)?
	if d.Cfg.Lambda.MaxConcurrentExecutions > 0 {
		svc.PutFunctionConcurrency(&lambda.PutFunctionConcurrencyInput{
			// FunctionName: awsString(arn),
			FunctionName:                 aws.String(d.Cfg.Lambda.FunctionName),
			ReservedConcurrentExecutions: aws.Int64(d.Cfg.Lambda.MaxConcurrentExecutions),
		})
	}
}

// AddLambdaInvokePermission will add permission to trigger Lambda (could be for a CloudWatch event rule or S3 bucket notification, etc.)
// Principal for CloudWatch event rules should be:
// "events.amazonaws.com"
// for S3 bucket notifications:
// "s3.amazonaws.com"
func (d *Deployer) AddLambdaInvokePermission(sourceArn string, principal string, statementID string) {
	// Both resources are created and the rule even looks like it should invoke the Lambda.
	// But it won't without this permission.
	// https://stackoverflow.com/questions/37571581/aws-cloudwatch-event-puttargets-not-adding-lambda-event-sources

	svc := lambda.New(d.AWSSession)

	_, err := svc.AddPermission(&lambda.AddPermissionInput{
		Action:       aws.String("lambda:InvokeFunction"),
		FunctionName: aws.String(d.Cfg.Lambda.FunctionName),
		Principal:    aws.String(principal),
		StatementId:  aws.String(statementID), // (any string should do, but must be unique per trigger)
		// EventSourceToken: aws.String("EventSourceToken"),
		// Qualifier:        aws.String("Qualifier"),
		// SourceAccount:    aws.String("SourceOwner"),
		SourceArn: aws.String(sourceArn),
	})
	if err != nil {
		// Ignore "already exists" errors, that's fine. No apparent way to look up permissions before making the add call?
		match, _ := regexp.MatchString("already exists", err.Error())
		if !match {
			fmt.Println("There was a problem setting permissions for the CloudWatch Rule to invoke the Lambda. Try again or go into AWS console and manually add the CloudWatch event rule trigger.")
			fmt.Println(err.Error())
		}
	}
}
