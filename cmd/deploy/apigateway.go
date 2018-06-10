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
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/fatih/color"
	swagger "github.com/tmaiaroto/aegis/apigateway"
	"github.com/tmaiaroto/aegis/cmd/config"
	"github.com/tmaiaroto/aegis/cmd/util"
)

// ImportAPI will import an API using Swagger
func (d *Deployer) ImportAPI(lambdaArn string) string {
	svc := apigateway.New(d.AWSSession)

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
		if *apisResp.Items[key].Name == d.Cfg.API.Name {
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
		Title:             d.Cfg.API.Name,
		LambdaURI:         swagger.GetLambdaURI(lambdaArn),
		ResourceTimeoutMs: d.Cfg.API.ResourceTimeoutMs,
		// BinaryMediaTypes: d.Cfg.API.BinaryMediaTypes,
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

// DeployAPI will create a stage and deploy the API
func (d *Deployer) DeployAPI(apiID string, stage config.DeploymentStage) string {
	svc := apigateway.New(d.AWSSession)

	// Must be one of: [58.2, 13.5, 28.4, 237, 0.5, 118, 6.1, 1.6]
	// TODO: Validate user input. Maybe round to nearest value
	if stage.CacheSize == "" {
		stage.CacheSize = apigateway.CacheClusterSize05
	}

	if stage.Cache {
		fmt.Printf("A cache is set for API responses, this will incur additional charges. Cache size is %sGB\n", stage.CacheSize)
	}

	// Figure out the stage variables, allowing for case sensitive keys
	stageVars := d.getStageVars(stage.Variables)

	_, err := svc.CreateDeployment(&apigateway.CreateDeploymentInput{
		RestApiId:           aws.String(apiID),      // Required
		StageName:           aws.String(stage.Name), // Required
		CacheClusterEnabled: aws.Bool(stage.Cache),
		CacheClusterSize:    aws.String(stage.CacheSize),
		Description:         aws.String(d.Cfg.API.Description),
		StageDescription:    aws.String(stage.Description),
		// Variables: stageVars,
		Variables: d.LookupSecretsForAPIGWStageVars(stageVars),
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
	buffer.WriteString(d.Cfg.AWS.Region)
	buffer.WriteString(".amazonaws.com/")
	buffer.WriteString(stage.Name)
	invokeURL := buffer.String()
	buffer.Reset()

	return invokeURL
}

// AddAPIPermission will add proper permissions to the API so that it can invoke the Lambda
func (d *Deployer) AddAPIPermission(apiID string, lambdaArn string) {
	// http://stackoverflow.com/questions/39905255/how-can-i-grant-permission-to-api-gateway-to-invoke-lambda-functions-through-clo
	// Glue together this weird SourceArn: arn:aws:execute-api:us-east-1:ACCOUNT_ID:API_ID/*/METHOD/ENDPOINT
	// Not sure if some API call can get it?
	accountID, region := util.GetAccountInfoFromLambdaArn(lambdaArn)

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

	svc := lambda.New(d.AWSSession)

	// There's no list permissions? So remove first and add.
	// _, err := svc.RemovePermission(&lambda.RemovePermissionInput{
	// 	FunctionName: aws.String("FunctionName"), // Required
	// 	StatementId:  aws.String("StatementId"),  // Required
	// 	Qualifier:    aws.String("Qualifier"),
	// })

	_, err := svc.AddPermission(&lambda.AddPermissionInput{
		Action:       aws.String("lambda:InvokeFunction"),           // Required
		FunctionName: aws.String(d.Cfg.Lambda.FunctionName),         // Required
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

// AddBinaryMediaTypes will update the API to specify valid binary media types
func (d *Deployer) AddBinaryMediaTypes(apiID string) {
	svc := apigateway.New(d.AWSSession)
	_, err := svc.UpdateRestApi(&apigateway.UpdateRestApiInput{
		RestApiId: aws.String(apiID), // Required
		PatchOperations: []*apigateway.PatchOperation{
			{
				Op: aws.String("add"),
				// TODO: Use configuration to set this...But that requires a function to escape and format this sring.
				// *~1* is */* which handles everything...Which could be enough...But maybe someone will want to only
				// accept specific media types? I don't know if there's any harm with this wildcard.
				// More info here: http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-payload-encodings-configure-with-control-service-api.html#api-gateway-payload-encodings-setup-with-api-set-encodings-map
				Path: aws.String("/binaryMediaTypes/*~1*"),
			},
		},
	})
	if err != nil {
		fmt.Println("There was a problem setting the binary media types for the API.")
	}
}

// UpdateAPI will update an API's settings that are not configured in the demployment/stage.
// There is no real need to update the resources or integrations of course, but things like
// the description, name, binary content types, etc. will need to be updated if changed.
// TODO: Unused. Maybe implement this.
func (d *Deployer) UpdateAPI(apiID string, lambdaArn string) {
	svc := apigateway.New(d.AWSSession)

	// Build Swagger
	swaggerDefinition, swaggerErr := swagger.NewSwagger(&swagger.SwaggerConfig{
		Title:             d.Cfg.API.Name,
		LambdaURI:         swagger.GetLambdaURI(lambdaArn),
		ResourceTimeoutMs: d.Cfg.API.ResourceTimeoutMs,
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

// getStageVars will Figure out how the stage.Variables are set. They can be string key value pairs OR the map can be
// multiple map[string]string in order to support case sensitive keys since viper will lowercase them from the YAML.
func (d *Deployer) getStageVars(variables map[string]interface{}) map[string]*string {
	stageVars := make(map[string]*string)

	for k, v := range variables {
		if v != nil {
			// If it's a string value
			if v, ok := v.(string); ok {
				stageVars[k] = &v
			}
			// If it's a map, it should be in the format of {"key": "foo", "value": "bar"}
			if v, ok := v.(map[string]interface{}); ok {
				// for example, v = map[value:<cognitoExample.ClientID> key:PoolID]
				if mKey, ok := v["key"]; ok {
					if key, ok := mKey.(string); ok {
						if mVal, ok := v["value"]; ok {
							if value, ok := mVal.(string); ok {
								stageVars[key] = &value
							}
						}
					}
				}
			}
		}
	}

	return stageVars
}
