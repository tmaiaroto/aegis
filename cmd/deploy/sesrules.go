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

package deploy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/fatih/color"
	"github.com/tmaiaroto/aegis/cmd/config"
	"github.com/tmaiaroto/aegis/cmd/util"
)

// AddSESRules will add SES rules from configuration
func (d *Deployer) AddSESRules() {
	svc := ses.New(d.AWSSession)

	var lastRule *string
	for _, rule := range d.Cfg.SESRules {
		// There are some required fields
		// The rule name must:
		//    * This value can only contain ASCII letters (a-z, A-Z), numbers (0-9),
		//    underscores (_), or dashes (-).
		//
		//    * Start and end with a letter or number.
		//
		//    * Contain less than 64 characters.
		if rule.RuleName != "" {

			// Rule set name
			// NOTE: Only one can be active at a time
			ruleSetName := aws.String("default-aegis-rule-set")
			if rule.RuleSet != "" {
				ruleSetName = aws.String(rule.RuleSet)
			}
			// It must be created if it doesn't already exist
			_, err := svc.CreateReceiptRuleSet(&ses.CreateReceiptRuleSetInput{
				RuleSetName: ruleSetName,
			})
			if err != nil {
				match, _ := regexp.MatchString("already exists", err.Error())
				if !match {
					fmt.Println("There was a problem creating the rule set.")
					fmt.Println(err)
				}
			} else {
				// Now set it active
				err := d.SetSESRuleSetActive(ruleSetName)
				if err != nil {
					fmt.Printf("There was a problem activating the rule set: %v\n", ruleSetName)
					fmt.Println("Check your AWS web console to ensure your rule set exists and is active. Only one rule set can be active at a time.")
				}
			}

			// This needs to be either Require or Optional, but configuration takes it as bool
			TLSPolicy := aws.String("Require")
			if !rule.RequireTLS {
				TLSPolicy = aws.String("Optional")
			}

			// If not defined; all recipients across all domains
			recipients := []*string{}
			if rule.Recipients != nil && len(rule.Recipients) > 0 {
				for _, recipient := range rule.Recipients {
					recipients = append(recipients, aws.String(recipient))
				}
			}

			// Build the Action
			// This needs to be Event or RequestResponse
			invocationType := aws.String("Event")
			if rule.InvocationType != "" {
				switch rule.InvocationType {
				case "Event", "event":
					invocationType = aws.String("Event")
				case "RequestResponse", "requestresponse", "requestResponse":
					invocationType = aws.String("RequestResponse")
				}
			}

			// Aegis only concerns itself with one action - the current Lambda
			// (at least for now)
			action := &ses.ReceiptAction{
				LambdaAction: &ses.LambdaAction{
					FunctionArn:    d.LambdaArn,
					InvocationType: invocationType,
				},
			}
			// This is optional
			if rule.SNSTopicArn != "" {
				action.LambdaAction.TopicArn = aws.String(rule.SNSTopicArn)
			}

			actions := []*ses.ReceiptAction{}
			actions = append(actions, action)

			// Optionally, e-mail messages can also be sent to an S3 bucket.
			// Note that when handling incoming SES events, only the raw e-mail is in the event.
			// That includes the headers, recipients, subject, etc. It does NOT include the body.
			// That's why it's often desired to store the messages in an S3 bucket.
			// That also means an S3ObjectRouter needs to be used in order to deal with e-mail contents.
			// The only config key needed here is the S3 bucket. All else is optional.
			if rule.S3Bucket != "" {
				// The S3 bucket needs a policy for SES to be able to put messages there.
				// In fact, this may need to be set first before this rule can even be saved.
				// Otherwise, the following error:
				// InvalidS3Configuration: Could not write to bucket: xxxxxx
				d.AddSESPolicyForS3Bucket(rule.S3Bucket)

				s3Action := &ses.ReceiptAction{
					S3Action: &ses.S3Action{
						BucketName: aws.String(rule.S3Bucket),
					},
				}
				// Optional
				if rule.S3ObjectKeyPrefix != "" {
					s3Action.S3Action.ObjectKeyPrefix = aws.String(rule.S3ObjectKeyPrefix)
				}
				// Optional. If messages to be encrypted using a specific KMS key
				// A default can be used: arn:aws:kms:REGION:ACCOUNT-ID-WITHOUT-HYPHENS:alias/aws/ses
				// (presumably what AWS web console does with its checkbox)
				if rule.S3EncryptMessage {
					KMSArn := ""
					// If a specific KMS key ARN was provided
					if rule.S3KMSKeyArn != "" {
						KMSArn = rule.S3KMSKeyArn
					} else {
						// Else, use the default
						lambdaArn := *d.LambdaArn
						accountID, region := util.GetAccountInfoFromLambdaArn(lambdaArn)
						var buffer bytes.Buffer
						buffer.WriteString("arn:aws:kms:")
						buffer.WriteString(region)
						buffer.WriteString(":")
						buffer.WriteString(accountID)
						buffer.WriteString(":alias/aws/ses")
						KMSArn = buffer.String()
						buffer.Reset()
					}
					s3Action.S3Action.KmsKeyArn = aws.String(KMSArn)
				}
				// Also optional. An SNS topic for the S3 action.
				if rule.S3SNSTopicArn != "" {
					s3Action.S3Action.TopicArn = aws.String(rule.S3SNSTopicArn)
				}

				actions = append(actions, s3Action)
			}

			// The rule to create (or update)
			receiptRule := &ses.ReceiptRule{
				Actions:     actions,
				Enabled:     aws.Bool(rule.Enabled),
				Name:        aws.String(rule.RuleName),
				Recipients:  recipients,
				ScanEnabled: aws.Bool(rule.ScanEnabled),
				TlsPolicy:   TLSPolicy,
			}

			// Create the rule with the single action
			_, err = svc.CreateReceiptRule(&ses.CreateReceiptRuleInput{
				After:       lastRule,
				Rule:        receiptRule,
				RuleSetName: ruleSetName,
			})

			if err != nil {
				match, _ := regexp.MatchString("already exists", err.Error())
				if !match {
					fmt.Println("There was a problem creating the SES Recipient Rule.")
					fmt.Println(err)
				} else {
					// Exists already? Update it.
					_, err := svc.UpdateReceiptRule(&ses.UpdateReceiptRuleInput{
						RuleSetName: ruleSetName,
						Rule:        receiptRule,
					})
					if err != nil {
						fmt.Println("There was a problem updating the SES Recipient Rule.")
						fmt.Println(err)
					}
				}
			} else {
				fmt.Printf("%v %v\n", "Added/updated SES Recipient Rule: ", color.GreenString(rule.RuleName))
			}

			// So the listed rules in configuration will be in order
			lastRule = aws.String(rule.RuleName)
		} else {
			fmt.Println("Missing required fields for SES Recipient Rule. You at least need a rule name.")
		}
	}
}

// AddSESPermission allows SES to invoke the Lambda
// See: https://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email-permissions.html
// Note that no permissions are required for SNS Topic unless it's outside the current account.
// The permissions for Lambda invocation is also by account, using `SourceAccount` unlike other
// some triggers like API Gateway which is by API GW ARN. So this only needs to be called once.
func (d *Deployer) AddSESPermission(lambdaArn *string) {
	if lambdaArn != nil {
		lambdaArnStr := *lambdaArn
		accountID, _ := util.GetAccountInfoFromLambdaArn(lambdaArnStr)

		svc := lambda.New(d.AWSSession)
		_, err := svc.AddPermission(&lambda.AddPermissionInput{
			Action:        aws.String("lambda:InvokeFunction"),
			FunctionName:  aws.String(d.Cfg.Lambda.FunctionName),
			Principal:     aws.String("ses.amazonaws.com"),
			StatementId:   aws.String("aegis-ses-rule-invoke-lambda"),
			SourceAccount: aws.String(accountID),
		})
		if err != nil {
			// Ignore "already exists" errors, that's fine. No apparent way to look up permissions before making the add call?
			match, _ := regexp.MatchString("already exists", err.Error())
			if !match {
				fmt.Println("There was a problem setting permissions for SES to invoke the Lambda. You may need to set this up manually.")
				fmt.Println(err.Error())
			}
		}
	}
}

// SetSESRuleSetActive will set a given rule set as the active set, only one can be active at a time with SES.
func (d *Deployer) SetSESRuleSetActive(ruleSetName *string) error {
	svc := ses.New(d.AWSSession)
	_, err := svc.SetActiveReceiptRuleSet(&ses.SetActiveReceiptRuleSetInput{
		RuleSetName: ruleSetName,
	})
	return err
}

// AddSESPolicyForS3Bucket will add a policy on the given S3 bucket to allow SES to store messages in it
// A policy allowing SES to put objects into S3 looks like this:
// {
// 	"Version": "2012-10-17",
// 	"Statement": [
// 		{
// 			"Sid": "AllowSESPuts",
// 			"Effect": "Allow",
// 			"Principal": {
// 				"Service": "ses.amazonaws.com"
// 			},
// 			"Action": "s3:PutObject",
// 			"Resource": "arn:aws:s3:::BUCKET-NAME/*",
// 			"Condition": {
// 				"StringEquals": {
// 					"aws:Referer": "AWSACCOUNTID"
// 				}
// 			}
// 		}
// 	]
// }
func (d *Deployer) AddSESPolicyForS3Bucket(bucketName string) error {
	if bucketName == "" {
		return errors.New("missing bucket name")
	}

	svc := s3.New(d.AWSSession)
	var err error

	// Get the current policy (if it exists)
	output, _ := svc.GetBucketPolicy(&s3.GetBucketPolicyInput{Bucket: aws.String(bucketName)})
	var policy config.S3BucketPolicy
	json.Unmarshal([]byte(aws.StringValue(output.Policy)), &policy)

	// Check to see if the policy statement already exists.
	statementExists := false
	if policy.Statement != nil {
		for _, statement := range policy.Statement {
			if statement.Sid == "AllowSESPuts" {
				statementExists = true
			}
		}
	}

	// If it doesn't exist on the bucket, add it.
	if !statementExists {
		lambdaArn := *d.LambdaArn
		accountID, _ := util.GetAccountInfoFromLambdaArn(lambdaArn)
		var buffer bytes.Buffer
		buffer.WriteString("arn:aws:s3:::")
		buffer.WriteString(bucketName)
		buffer.WriteString("/*")
		resource := buffer.String()
		buffer.Reset()

		allowSESStatement := config.PolicyStatement{
			Sid:    "AllowSESPuts",
			Effect: "Allow",
			Principal: config.StatementPrincipal{
				Service: "ses.amazonaws.com",
			},
			Action:   []string{"s3:PutObject"},
			Resource: []string{resource},
			Condition: config.StatementCondition{
				StringEquals: map[string]string{"aws:Referer": accountID},
			},
		}
		if policy.Statement == nil {
			policy.Statement = make([]config.PolicyStatement, 0)
		}
		policy.Statement = append(policy.Statement, allowSESStatement)

		// If this policy was actually brand new, then we'll have the struct with some default empty fields
		if policy.Version == "" {
			policy.Version = "2012-10-17"
		}
		if policy.ID == "" {
			policy.ID = "default"
		}

		policyBytes, err := json.Marshal(policy)
		if err == nil {
			// Put the policy on the S3 bucket
			_, err = svc.PutBucketPolicy(&s3.PutBucketPolicyInput{Bucket: aws.String(bucketName), Policy: aws.String(string(policyBytes))})
		}
		if err != nil {
			fmt.Println("Error setting the S3 bucket policy so that SES can store messages into it. Try setting it manually.")
			fmt.Printf("Bucket name: %v\n", color.RedString(bucketName))
			fmt.Println(err)
		} else {
			fmt.Printf("Updated S3 bucket policy so that SES can store messages to: %v\n", color.GreenString(bucketName))
		}
	}

	return err
}
