package framework

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mitchellh/mapstructure"
)

// Handlers defines a set of Aegis framework Lambda handlers
type Handlers struct {
	Router         *Router
	Tasker         *Tasker
	RPCRouter      *RPCRouter
	DefaultHandler DefaultHandler
}

// DefaultHandler is used when the message type can't be identified as anything else, completely optional to use
type DefaultHandler func(context.Context, *map[string]interface{}) (interface{}, error)

// getType will determine which type of event is being sent
func getType(evt map[string]interface{}) string {
	// log.Println("evt:", evt)

	// if APIGatewayProxyRequest
	if keyInMap("httpMethod", evt) && keyInMap("path", evt) {
		return "APIGatewayProxyRequest"
	}

	// TODO: Look into this again later...It turns out that scheduled events from CloudWatch
	// will just send the static input JSON over alone by itself as a map. There will be
	// no proper AWS Lambda event type struct. Context also apparently provides no insight
	// that the Lambda was invoked by CloudWatch either.
	// So it's up to Aegis to handle by convention.
	//
	// if CloudWatchEvent or AutoScalingEvent
	// if keyInMap("DetailType", evt) && keyInMap("Source", evt) && keyInMap("Detail", evt) {
	// 	switch t := evt["Detail"].(type) {
	// 	case string:
	// 		return "AutoScalingEvent"
	// 	case json.RawMessage:
	// 		return "CloudWatchEvent"
	// 	default:
	// 		_ = t
	// 		return ""
	// 	}
	// }

	// The convention will be that tasks are named with a `_taskName` key.
	// This is known as an "AegisTask" and gets handled by Tasker.
	if keyInMap("_taskName", evt) {
		return "AegisTask"
	}

	if keyInMap("_rpcName", evt) {
		return "AegisRPC"
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
	switch getType(evt) {
	case "APIGatewayProxyRequest":
		var e APIGatewayProxyRequest
		// This mapstructure package does use reflection.
		// An alternative to this would be to go back to JSON then to struct.
		// TODO: consider this. does it really matter? is one way faster?
		err = mapstructure.Decode(evt, &e)
		if err == nil {
			return h.Router.LambdaHandler(ctx, e)
		}
	case "AegisTask":
		// Task handlers have no return
		// Tasker takes a simple map[string]interface{} - not a struct (like some other events).
		h.Tasker.LambdaHandler(ctx, evt)
		return nil, nil
	case "AegisRPC":
		return h.RPCRouter.LambdaHandler(ctx, evt)
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
