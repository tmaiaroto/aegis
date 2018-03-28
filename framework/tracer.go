package framework

import (
	"context"

	"github.com/aws/aws-xray-sdk-go/xray"
)

// AWSClientTracer is an alias of xray.AWS. Overwrite if framework is to use a different client tracer.
var AWSClientTracer = xray.AWS

// TraceStrategy interface uses XRay by default. Any struct implementing this interface can be used instead.
type TraceStrategy struct {
	Annotations map[string]interface{}
	Metadata    map[string]interface{}
}

// Capture traces the provided synchronous function by using XRay Cpature() which puts a beginning and closing subsegment around its execution
func (t *TraceStrategy) Capture(ctx context.Context, name string, fn func(context.Context) error) (err error) {
	return xray.Capture(ctx, name, fn)
}

// CaptureAsync traces an arbitrary code segment within a goroutine by using XRay CaptureAsync()
func CaptureAsync(ctx context.Context, name string, fn func(context.Context) error) {
	xray.CaptureAsync(ctx, name, fn)
}

// AddAnnotations will add annotations to the trace if configured
func (t *TraceStrategy) AddAnnotations(ctx context.Context) {
	if t.Annotations != nil {
		for k, v := range t.Annotations {
			xray.AddAnnotation(ctx, k, v)
		}
	}
}

// AddMetadata will add meta data to the trace if configured
func (t *TraceStrategy) AddMetadata(ctx context.Context) {
	if t.Metadata != nil {
		for k, v := range t.Metadata {
			xray.AddMetadata(ctx, k, v)
		}
	}
}
