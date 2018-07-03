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
	"log"

	"github.com/aws/aws-sdk-go/service/sqs"
)

// AddQueues will add SQS queues, if they don't exist, and apply trigger to the lambda
// NOTE: Apparently only one SQS can be associated with a Lambda at a time (at least the CLI doesn't let you trigger multiple)
// Though a Lambda can be triggered by multiple different queues
func (d *Deployer) AddQueues() {
	for _, queue := range d.Cfg.Queues {
		sqsClient := sqs.New(d.AWSSession)

		log.Println(queue)
		log.Println(sqsClient)

		// TODO: Unsure if the AWS Go SDK has the ability to set these triggers for SQS yet

		// Look up ARN, if not, create queue and get arn.
		// sqsArn := d.createOrUpdateQueue(queue)

		// Add trigger
		// ruleArn := d.addCloudWatchEventRuleForLambda(task, d.LambdaArn)

		// After the rule is added, we have to give it permissions to invoke the Lambda (even though we gave it a target)
		// d.AddLambdaInvokePermission(ruleArn, "events.amazonaws.com", "aegis-cloudwatch-rule-invoke-lambda")
	}
}
