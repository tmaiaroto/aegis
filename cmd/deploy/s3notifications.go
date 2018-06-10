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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fatih/color"
)

// AddS3BucketNotifications loops the buckets in configuration and sets appropriate notifications to trigger the Lambda
func (d *Deployer) AddS3BucketNotifications() {
	if d.Cfg.BucketTriggers != nil {
		for _, trigger := range d.Cfg.BucketTriggers {
			d.addS3BucketNotification(trigger.Bucket, trigger.EventNames, trigger.Filters, trigger.Disabled)
		}
	}
}

// addS3BucketNotification will add a notification to trigger the Lambda for S3 bucket events
// see: https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
func (d *Deployer) addS3BucketNotification(bucket *string, eventNames []*string, filterRules []*s3.FilterRule, disabled bool) {
	svc := s3.New(d.AWSSession)

	// log.Println("bucket", bucket)
	// bName := *bucket
	// log.Println("bucket name", bName)
	// log.Println(eventNames)
	// log.Println(filterRules)

	// If disabled was passed true, then just remove the events.
	// There is actually no way to disable from SDK.
	if disabled {
		eventNames = make([]*string, 0)
	}
	// S3 bucket ARN is always this format.
	bucketArn := "arn:aws:s3:::" + aws.StringValue(bucket)
	statementID := "aeigs_invoke_lambda_with" + aws.StringValue(bucket)
	d.AddLambdaInvokePermission(bucketArn, "s3.amazonaws.com", statementID)

	_, err := svc.PutBucketNotificationConfiguration(&s3.PutBucketNotificationConfigurationInput{
		Bucket: bucket,
		NotificationConfiguration: &s3.NotificationConfiguration{
			LambdaFunctionConfigurations: []*s3.LambdaFunctionConfiguration{
				&s3.LambdaFunctionConfiguration{
					LambdaFunctionArn: d.LambdaArn,
					Events:            eventNames,
					Filter: &s3.NotificationConfigurationFilter{
						Key: &s3.KeyFilter{
							FilterRules: filterRules,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println("There was an error setting the S3 notification trigger.")
		fmt.Println(err)
	} else {
		bucketName := *bucket
		fmt.Printf("%v %v\n", "Added/updated S3 notification trigger for bucket:", color.GreenString(bucketName))
	}
}

// Example JSON for event pattern
// {
// 	"source": [
// 	  "aws.s3"
// 	],
// 	"detail-type": [
// 	  "AWS API Call via CloudTrail"
// 	],
// 	"detail": {
// 	  "eventSource": [
// 		"s3.amazonaws.com"
// 	  ],
// 	  "eventName": [
// 		"ListObjects",
// 		"ListObjectVersions",
// 		"PutObject",
// 		"GetObject",
// 		"HeadObject",
// 		"CopyObject",
// 		"GetObjectAcl",
// 		"PutObjectAcl",
// 		"CreateMultipartUpload",
// 		"ListParts",
// 		"UploadPart",
// 		"CompleteMultipartUpload",
// 		"AbortMultipartUpload",
// 		"UploadPartCopy",
// 		"RestoreObject",
// 		"DeleteObject",
// 		"DeleteObjects",
// 		"GetObjectTorrent"
// 	  ],
// 	  "requestParameters": {
// 		"bucketName": [
// 		  ""
// 		]
// 	  }
// 	}
//   }

// This ended up being wrong actually. You don't add a CloudWatch event rule.
// Though that was rather misleading in the web console because it looked like it should work.
// Then adding permission for the rule to the Lambda also looked pretty legit.
// But really, it's an S3 trigger you add with the designer it shows it's a completely different thing.
// It's an S3 notification. Not a CloudWatch event. Interesting.
// Though the CloudWatch event JSON work done below was a bit time consuming and I wanted to keep it somewhere.
// I have a feeling I'll use it again some day.
// move to gist.

// Name based on buckets
// hashStr := strconv.FormatUint(hash, 10)
// ruleName := strings.ToLower(d.Cfg.Lambda.FunctionName + "_s3_events")

// rulePattern := config.CloudWatchRuleEventPattern{
// 	Source:     []string{"aws.s3"},
// 	DetailType: []string{"AWS API Call via CloudTrail"},
// 	Detail: map[string]interface{}{
// 		"requestParameters": map[string][]string{
// 			"bucketName": buckets,
// 		},
// 		"eventName":   eventNames,
// 		"eventSource": []string{"s3.amazonaws.com"},
// 	},
// }

// b, err := json.Marshal(rulePattern)
// if err == nil {
// 	jsonEventPattern := string(b)

// 	svc := cloudwatchevents.New(d.AWSSession)
// 	ruleOutput, err := svc.PutRule(&cloudwatchevents.PutRuleInput{
// 		Description: aws.String("S3 Object event triggers"),
// 		// Again, name is all lowercase, filename without extension: <function name>_<file name>
// 		// ie. for an example.json file, an event ARN/ID like: arn:aws:events:us-east-1:1234567890:rule/aegis_aegis_example
// 		Name: aws.String(ruleName),
// 		// IAM Role (either defined in aegix.yml or was set after creating role - either way, it should be on cfg by now)
// 		RoleArn: aws.String(d.Cfg.Lambda.Role),
// 		// ENABLED or DISABLED
// 		// In the Task JSON, the field is "disabled" so that when marshaling it defaults to false.
// 		// This way it's optional in the JSON which makes it enabled by default.
// 		State: aws.String(state),
// 		// Here's the fun part
// 		EventPattern: aws.String(jsonEventPattern),
// 	})
// 	if err != nil {
// 		fmt.Println("There was a problem creating a CloudWatch Event Rule.")
// 		fmt.Println(err)
// 	} else {
// 		ruleArn = *ruleOutput.RuleArn
// 	}

// 	// Add the Lambda ARN as the target
// 	_, err = svc.PutTargets(&cloudwatchevents.PutTargetsInput{
// 		Rule: aws.String(ruleName),
// 		Targets: []*cloudwatchevents.Target{
// 			&cloudwatchevents.Target{
// 				Arn: d.LambdaArn,
// 				Id:  aws.String(ruleName),
// 			},
// 		},
// 	})

// 	if err != nil {
// 		fmt.Println("There was an error setting the CloudWatch Event Rule Target (Lambda function).")
// 		fmt.Println(err)
// 	} else {
// 		fmt.Printf("%v %v\n", "Added/updated S3 Object Event Rule:", color.GreenString(ruleName))
// 	}
// } else {
// 	fmt.Println("There was a problem creating a CloudWatch Event Rule.")
// 	fmt.Println(err)
// }

// Also not right ... it's not a bucket policy. It's a permission.
// Lots of bad information out there on the internet.
// output, _ := svc.GetBucketPolicy(&s3.GetBucketPolicyInput{Bucket: bucket})
// 	var policy config.S3BucketPolicy
// 	json.Unmarshal([]byte(aws.StringValue(output.Policy)), &policy)

// 	needsPermission := true
// 	if policy.Statement != nil {
// 		for _, statement := range policy.Statement {
// 			if statement.Resource == aws.StringValue(d.LambdaArn) {
// 				for _, action := range statement.Action {
// 					if action == "lambda:InvokeFunction" || action == "lambda:*" {
// 						needsPermission = false
// 					}
// 				}
// 			}
// 		}
// 	}

// 	if needsPermission {
// 		if policy.Version == "" {
// 			policy.Version = "2012-10-17"
// 		}
// 		if policy.Statement == nil {
// 			policy.Statement = make([]config.PolicyStatement, 0)
// 		}
// 		policy.Statement = append(policy.Statement, config.PolicyStatement{
// 			Sid:    "aegis_s3_lambda_trigger",
// 			Effect: "Allow",
// 			Principal: config.StatementPrincipal{
// 				Service: "s3.amazonaws.com",
// 			},
// 			Action:   []string{"lambda:InvokeFunction"},
// 			Resource: aws.StringValue(d.LambdaArn),
// 		})

// 		policyStr, err := json.Marshal(policy)
// 		if err == nil {
// 			// Put the policy on the S3 bucket
// 			_, err = svc.PutBucketPolicy(&s3.PutBucketPolicyInput{Bucket: bucket, Policy: aws.String(string(policyStr))})
// 		}
// 		if err != nil {
// 			fmt.Println("Error setting the S3 bucket policy so that it can invoke the Lambda. Try setting it manually.")
// 			fmt.Println(err)
// 		}
// 	}
