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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// retrievedSecrets holds any secrets retrieved during deploy to prevent additional unnecessary requests
// for other keys within the same secret
var retrievedSecrets map[string]interface{}

// Secrets are added by existing deploy functions.
// For Lambda env vars, it's added by the Lambda create/update input struct.
// Using:
// d.Cfg.Lambda.EnvironmentVariables
//
// What we need this secrets.go file to do is allow lookups for variables.
// Essentially a function that can be given this `d.Cfg.Lambda.EnvironmentVariables`
// that then loops it and looks for <scretName.keyName> values and decodes/retrieves them
// from AWS Secrets Manager. Then it returns the same struct/map value, but with replaced/decoded
// variables to use instead.
//
// The same then goes for API Gateway stage variables.
// ALTHOUGH: We need to be careful about special characters there.
// Either ! or @ is invalid I think. Or both? There's a few.

// LookupSecretsForLambdaEnvVars will look up variables from AWS Secrets Manager for use with Lambda environment variables.
func (d *Deployer) LookupSecretsForLambdaEnvVars(vars map[string]*string) map[string]*string {
	// There are limits on Lambda environment variables as well...But just the names.
	// https://docs.aws.amazon.com/lambda/latest/dg/env_variables.html
	// Names Must start with letters [a-zA-Z].
	// Can only contain alphanumeric characters and underscores ([a-zA-Z0-9_].
	// https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html#lambda-environment-variables
	// The actual values don't seem to have any real limits (maybe size/length?).
	// So they can contain special characters. Just use LookupSecretsForVars() as is.
	return d.LookupSecretsForVars(vars)
}

// LookupSecretsForAPIGWStageVars will look up variables from AWS Secrets Manager for use with API Gateway stage variables.
// API Gateway stage variable limitations:
// Variable names can have alphanumeric and underscore characters, and the values must match [A-Za-z0-9-._~:/?#&=,]+.
func (d *Deployer) LookupSecretsForAPIGWStageVars(vars map[string]*string) map[string]*string {
	// rValidVariableName, _ := regexp.Compile("^[A-z\\_]*$")
	// could do something like...
	// rValidVariableName.MatchString(vars[k])
	// ...but that may be best left done elsewhere.
	stageVars := d.LookupSecretsForVars(vars)
	rAPIGWStageVar, _ := regexp.Compile("^[A-Za-z0-9-._~:/?#&=,]*$")
	for k, v := range stageVars {
		value := *v
		if value != "" && !rAPIGWStageVar.MatchString(value) {
			stageVars[k] = aws.String(base64.StdEncoding.EncodeToString([]byte(value)))
		}
	}
	return stageVars
}

// LookupSecretsForVars will look up variables from AWS Secrets Manager, replacing values in a given map.
func (d *Deployer) LookupSecretsForVars(vars map[string]*string) map[string]*string {
	r, _ := regexp.Compile("\\<(.*)\\.(.*)\\>")
	for k, v := range vars {
		variable := *v
		matches := r.FindStringSubmatch(variable)
		// If proper format, should return ["<secretName.secretKey>" "secretName" "secretKey"]
		if len(matches) == 3 {
			vars[k] = aws.String(d.GetSecretsKeyValue(matches[1], matches[2]))
		}
	}
	return vars
}

// getSecret makes the actaul AWS SDK call to get the secret and keeps it on `retrievedSecrets`
// for future use during deployment.
func (d *Deployer) getSecret(secretName string) error {
	svc := secretsmanager.New(d.AWSSession)
	var err error

	if retrievedSecrets == nil {
		retrievedSecrets = make(map[string]interface{})
	}

	if _, ok := retrievedSecrets[secretName]; ok {
		return err
	}
	secret, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		// Can be either ARN or friendly name (fortunately)
		SecretId: aws.String(secretName),
	})

	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			fmt.Println("Secret not found.")
		} else {
			fmt.Println("Could not read secret.", err)
		}
		// If error is returned here, there could be future look ups for the values.
		// Whereas the code below will set an empty map[string]interface{}
		// that will be used by subsequent calls to GetSecretsKeyValue().
		// So one error will be displayed about the secret not being found
		// and one SDK call. Multiple errors displayed then about additional
		// keys not being found for this secret that couldn't be found or read.
		// return err
	}

	var kvE map[string]interface{}
	err = json.Unmarshal([]byte(aws.StringValue(secret.SecretString)), &kvE)
	retrievedSecrets[secretName] = kvE

	return err
}

// GetSecretsKeyValue will look up a secret from AWS Secrets Manager
func (d *Deployer) GetSecretsKeyValue(secretName string, keyName string) string {
	keyValue := ""

	err := d.getSecret(secretName)
	if err == nil {
		if secret, ok := retrievedSecrets[secretName]; ok {
			secretVal := secret.(map[string]interface{})
			if val, ok := secretVal[keyName]; ok {
				// Ultimately, all values in Lambda environment variables or API Gateway stage variables are strings.
				// It will be up to the application/developer to parse these values. Whether they put them in to the
				// environment variables or stage variables manually or if aegis did upon deployment.
				// keyValue = val.(string)
				// Always take the vlaue as a string. It may be able to technically be a boolean or int, etc.
				// Since it's JSON I believe...But we're after the string representation of it.
				keyValue = fmt.Sprintf("%v", val)
			}
		} else {
			fmt.Println("That key does not exist for " + secretName + ".")
		}
	}

	return keyValue
}
