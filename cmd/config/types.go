// various types are held here for re-use

package config

// CloudWatchRuleEventPattern defines an event pattern (ultimately sent as JSON)
type CloudWatchRuleEventPattern struct {
	Source     []string               `json:"source"`
	DetailType []string               `json:"detail-type"`
	Detail     map[string]interface{} `json:"detail"`
}

// S3BucketPolicy defines a generic bucket policy
// {
// 	"Version": "2012-10-17",
// 	"Id": "default",
// 	"Statement": [
// 	  {
// 		"Sid": "<optional>",
// 		"Effect": "Allow",
// 		"Principal": {
// 		  "Service": "s3.amazonaws.com"
// 		},
// 		"Action": "lambda:InvokeFunction",
// 		"Resource": "<ArnToYourFunction>",
// 		"Condition": {
// 		  "StringEquals": {
// 			"AWS:SourceAccount": "<YourAccountId>"
// 		  },
// 		  "ArnLike": {
// 			"AWS:SourceArn": "arn:aws:s3:::<YourBucketName>"
// 		  }
// 		}
// 	  }
// 	]
// }
type S3BucketPolicy struct {
	Version   string            `json:"Version"`
	ID        string            `json:"Id"`
	Statement []PolicyStatement `json:"Statement"`
}

// PolicyStatement defines a generic AWS policy statement
type PolicyStatement struct {
	Sid       string             `json:"Sid"`
	Effect    string             `json:"Effect"`
	Principal StatementPrincipal `json:"Principal"`
	Action    []string           `json:"Action"`
	Resource  string             `json:"Resource"`
	// Condition TODO? Not yet used by aegis
}

// StatementPrincipal defines a generic AWS policy statement principal (TODO: see what else there is besides Service in here)
type StatementPrincipal struct {
	Service string `json:"Service"`
}
