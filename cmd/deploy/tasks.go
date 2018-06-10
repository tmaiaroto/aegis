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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/fatih/color"
	"github.com/tdewolff/minify"
	mJson "github.com/tdewolff/minify/json"
	"github.com/tmaiaroto/aegis/cmd/config"
)

// AddTasks will add CloudWatch event rules to trigger the Lambda on set intervals with JSON messages from a `tasks` directory
func (d *Deployer) AddTasks() {
	for _, task := range d.getTasks() {
		ruleArn := d.addCloudWatchEventRuleForLambda(task, d.LambdaArn)
		// After the rule is added, we have to give it permissions to invoke the Lambda (even though we gave it a target)
		d.AddLambdaInvokePermission(ruleArn, "events.amazonaws.com", "aegis-cloudwatch-rule-invoke-lambda")
	}
}

// getTasks will scan a `tasks` directory looking for JSON files (this is where all tasks should be kept)
func (d *Deployer) getTasks() []*config.Task {
	var tasks []*config.Task

	// Don't proceed if the folder doesn't even exist.
	// log.Println("Looking for tasks in:", d.TasksPath)
	_, err := os.Stat(d.TasksPath)
	if os.IsNotExist(err) {
		return tasks
	}

	// Load the task definition files.
	f, err := os.Open(d.TasksPath)
	if err != nil {
		log.Printf("error opening tasks path: %s", err)
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	for _, file := range files {
		if file.Mode().IsRegular() {
			if strings.ToLower(filepath.Ext(file.Name())) == ".json" {
				fp, err := filepath.EvalSymlinks(d.TasksPath + "/" + file.Name())
				if err == nil {
					raw, err := ioutil.ReadFile(fp)
					if err == nil {
						var t config.Task
						// The scheduled task file (a JSON file) should define most everything needed.
						// The input, schedule, etc.
						json.Unmarshal(raw, &t)
						// Set a name based on the function name and file path.
						// This makes it easier to update for future deploys.
						// Do not allow a name override (for now - makes Tasker name matching easier, more conventional)
						filename := file.Name()
						extension := filepath.Ext(filename)
						name := filename[0 : len(filename)-len(extension)]
						t.Name = strings.ToLower(d.Cfg.Lambda.FunctionName + "_" + name)
						tasks = append(tasks, &t)
					}
				}
			}
		}
	}

	return tasks
}

// addCloudWatchEventRuleForLambda will add CloudWatch Event Rule for triggering the Lambda on a schedule with input
func (d *Deployer) addCloudWatchEventRuleForLambda(t *config.Task, lambdaArn *string) string {
	state := "ENABLED"
	if t.Disabled {
		state = "DISABLED"
	}
	ruleArn := ""
	// TODO: validation and log out errors/warnings?
	if t.Schedule != "" {
		svc := cloudwatchevents.New(d.AWSSession)
		ruleOutput, err := svc.PutRule(&cloudwatchevents.PutRuleInput{
			Description: aws.String(t.Description),
			// Again, name is all lowercase, filename without extension: <function name>_<file name>
			// ie. for an example.json file, an event ARN/ID like: arn:aws:events:us-east-1:1234567890:rule/aegis_aegis_example
			Name: aws.String(t.Name),
			// IAM Role (either defined in aegix.yml or was set after creating role - either way, it should be on cfg by now)
			RoleArn: aws.String(d.Cfg.Lambda.Role),
			// For example, "cron(0 20 * * ? *)" or "rate(5 minutes)"
			ScheduleExpression: aws.String(t.Schedule),
			// ENABLED or DISABLED
			// In the Task JSON, the field is "disabled" so that when marshaling it defaults to false.
			// This way it's optional in the JSON which makes it enabled by default.
			State: aws.String(state),
		})
		if err != nil {
			fmt.Println("There was a problem creating a CloudWatch Event Rule.")
			fmt.Println(err)
		} else {
			ruleArn = *ruleOutput.RuleArn
		}

		jsonInput, _ := t.Input.MarshalJSON()
		// Minify the JSON, trim whitespace, etc.
		m := minify.New()
		m.AddFuncRegexp(regexp.MustCompile("json$"), mJson.Minify)
		jsonBytes, _ := m.Bytes("json", jsonInput)

		var buffer bytes.Buffer
		buffer.WriteString(string(jsonBytes))
		inputTask := buffer.String()
		buffer.Reset()

		// Add the Lambda ARN as the target
		_, err = svc.PutTargets(&cloudwatchevents.PutTargetsInput{
			Rule: aws.String(t.Name),
			Targets: []*cloudwatchevents.Target{
				&cloudwatchevents.Target{
					Arn:   lambdaArn,
					Id:    aws.String(t.Name),
					Input: aws.String(inputTask),
				},
			},
		})

		if err != nil {
			fmt.Println("There was an error setting the CloudWatch Event Rule Target (Lambda function).")
			fmt.Println(err)
		} else {
			fmt.Printf("%v %v\n", "Added/updated Task:", color.GreenString(t.Name))
		}
	}

	return ruleArn
}
