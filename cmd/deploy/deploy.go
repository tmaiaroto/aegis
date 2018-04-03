// package deploy is purely for organization, the deploy.go command file was getting absurdly long
package deploy

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/tmaiaroto/aegis/cmd/config"
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
