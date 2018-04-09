package framework

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mitchellh/mapstructure"
)

// Handlers defines a set of Aegis framework Lambda handlers
type Handlers struct {
	Router         *Router
	Tasker         *Tasker
	RPCRouter      *RPCRouter
	S3ObjectRouter *S3ObjectRouter
	CognitoRouter  *CognitoRouter
	DefaultHandler DefaultHandler
}

// DefaultHandler is used when the message type can't be identified as anything else, completely optional to use
type DefaultHandler func(context.Context, *map[string]interface{}) (interface{}, error)

// getType will determine which type of event is being sent
func getType(evt map[string]interface{}) string {
	// if APIGatewayProxyRequest
	if keyInMap("httpMethod", evt) && keyInMap("path", evt) {
		return "APIGatewayProxyRequest"
	}

	// if S3Event
	if keyInMap("Records", evt) {
		records := evt["Records"].([]interface{})
		if len(records) > 0 {
			if keyInMap("s3", records[0].(map[string]interface{})) {
				return "S3Event"
			}
		}
	}

	// The convention will be that tasks are named with a `_taskName` key.
	// This is known as an "AegisTask" and gets handled by Tasker.
	if keyInMap("_taskName", evt) {
		return "AegisTask"
	}

	if keyInMap("_rpcName", evt) {
		return "AegisRPC"
	}

	// if Cognito trigger
	if keyInMap("userPoolId", evt) && keyInMap("triggerSource", evt) {
		return "CognitoTrigger"
	}

	// if CognitoEvent
	if keyInMap("identityPoolId", evt) && keyInMap("datasetRecords", evt) {
		return "CognitoEvent"
	}

	return ""
}

// keyInMap will simply check for the existence of a key in a given map
func keyInMap(k string, m map[string]interface{}) bool {
	if _, ok := m[k]; ok {
		return true
	}
	return false
}

// eventHandler is a general handler that accepts an interface and determines which hanlder to use based on the event.
// See: https://godoc.org/github.com/aws/aws-lambda-go/lambda#Start
func (h *Handlers) eventHandler(ctx context.Context, evt map[string]interface{}) (interface{}, error) {
	// log.Println("Determining type of event for:", evt)

	var err error
	// TODO: This isn't exactly reflection, it's a map.
	// But we do need to look at the signature to make a determination.
	evtType := getType(evt)
	log.Println("Incoming Lambda event type: ", evtType)
	switch evtType {
	case "APIGatewayProxyRequest":
		var e APIGatewayProxyRequest
		// The event contains no time/date, should decode just fine
		err = mapstructure.Decode(evt, &e)
		if err == nil {
			return h.Router.LambdaHandler(ctx, e)
		}
		log.Println("Could not decode APIGatewayProxyRequest event", err)
	case "AegisTask":
		// Task handlers have no return
		// Tasker takes a simple map[string]interface{} - not a struct (like some other events).
		h.Tasker.LambdaHandler(ctx, evt)
		return nil, nil
	case "AegisRPC":
		return h.RPCRouter.LambdaHandler(ctx, evt)
	case "S3Event":
		var e S3Event
		decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			// Event time format: 2018-04-02T17:09:32.273Z
			// mapstructure does not handle string to time.Time automatically so we need to use a hook
			DecodeHook: mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			Result:     &e,
		})
		// decodeErr := mapstructure.Decode(evt, &e)
		decodeErr := decoder.Decode(evt)
		if decodeErr == nil {
			err = h.S3ObjectRouter.LambdaHandler(ctx, e)
		} else {
			log.Println("Could not decode S3Event", decodeErr)
		}
	case "CognitoTrigger":
		// There's so many different formats here, routing for each is a bit silly.
		// So send map[string]interface{}
		// The handler itself can unmarshal using structs found in cognito_trigger_types.go
		return h.CognitoRouter.LambdaHandler(ctx, evt)
	default:
		log.Println("Could not determine Lambda event type, using DefaultHandler.")
		// If a default handler is not set, return an error about it.
		// It's essentially an unhandled Lambda invocation at this point.
		if h.DefaultHandler == nil {
			h.DefaultHandler = func(context.Context, *map[string]interface{}) (interface{}, error) {
				return nil, errors.New("unhandled event")
			}
		}
		return h.DefaultHandler(ctx, &evt)
	}

	if err != nil {
		log.Println(err)
	}
	return nil, err
}

// Listen will start a general listener which determines the proper handler to used based on incoming events
func (h *Handlers) Listen() {
	lambda.Start(h.eventHandler)
}
