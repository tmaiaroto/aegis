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

// package config allows re-use of the config struct
package config

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/s3"
)

// DeploymentConfig holds the AWS Lambda configuration
type DeploymentConfig struct {
	App struct {
		Name           string
		KeepBuildFiles bool
		BuildFileName  string
	}
	AWS struct {
		Region          string
		Profile         string
		AccessKeyID     string
		SecretAccessKey string
	}
	Lambda struct {
		Wrapper              string
		Runtime              string
		Handler              string
		FunctionName         string
		Alias                string
		Description          string
		MemorySize           int64
		Role                 string
		Timeout              int64
		SourceZip            string
		EnvironmentVariables map[string]*string
		KMSKeyArn            string
		VPC                  struct {
			SecurityGroups []string
			Subnets        []string
		}
		TraceMode               string
		MaxConcurrentExecutions int64
	}
	API struct {
		Name              string
		Description       string
		Cache             bool
		CacheSize         string
		Stages            map[string]DeploymentStage
		ResourceTimeoutMs int
		BinaryMediaTypes  []*string
	}
	BucketTriggers []BucketTrigger
}

// DeploymentStage defines an API Gateway stage and holds configuration options for it
type DeploymentStage struct {
	Name        string
	Description string
	Variables   map[string]interface{}
	Cache       bool
	CacheSize   string
}

// Task defines options for a CloudWatch event rule (scheduled task)
type Task struct {
	Schedule    string          `json:"schedule"`
	Input       json.RawMessage `json:"input"`
	Disabled    bool            `json:"disabled"`
	Description string          `json:"description"`
	Name        string          `json:"-"` // Do not allow names to be set by JSON files (for now)
}

// BucketTrigger defines options for S3 bucket notifications
type BucketTrigger struct {
	Bucket     *string
	Filters    []*s3.FilterRule
	EventNames []*string
	Disabled   bool
}

// bucketTriggers:
//   - bucket: aegis-incoming
//     filters:
//       - name: suffix
//         value: png
//       - name: prefix
//         value: filename/or/path
//     eventNames:
//       - s3:ObjectCreated:*
//       - s3:ObjectRemoved:*
//       # ... there's a few and there's wildcards, see:
//       # https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
//     disabled: false
