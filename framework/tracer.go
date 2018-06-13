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

package framework

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// TraceStrategy interface allows for customized tracing (AWS X-Ray is Aegis' default).
//
// You'll notice a lot of interface{} values here, that's because we don't know what the segment or data will be.
// In the case of AWS X-Ray the segment and subsegment values are both a Segment struct.
// Recorded data could be annotations, metadata, and more.
//
// The strategy supports the concept of opening and closing segments and subsegments, recording data, as well
// as capturing (synchronously and asynchronously) or "wrapping" function calls to be traced. Somewhere.
// It also concerns itself with context of course.
type TraceStrategy interface {
	// Record sends data to a trace service (either immediately or sometimes upon Capture() or CaptureAsync(), it depends on the strategy)
	Record(dataType string, data interface{})
	// BeginSegment starts a new trace segment (should be fairly common for anything that traces, though can be implemented to do nothing)
	BeginSegment(ctx context.Context, name string) (context.Context, interface{})
	// BeginSubsegment starts a new subsegment if applicable
	BeginSubsegment(ctx context.Context, name string) (context.Context, interface{})
	// CloseSegment will close a segment
	CloseSegment(interface{}, error)
	// CloseSubsegment will close a subsegment
	CloseSubsegment(interface{}, error)
	// Capture traces the provided synchronous function
	Capture(ctx context.Context, name string, fn func(context.Context) error) error
	// CaptureAsync traces an arbitrary code segment within a goroutine
	CaptureAsync(ctx context.Context, name string, fn func(context.Context) error)
}

// XRayTraceStrategy interface uses AWS X-Ray
type XRayTraceStrategy struct {
	Annotation        map[string]interface{}
	Metadata          map[string]interface{}
	NamespaceMetadata map[string]map[string]interface{}
	Error             error
	AWSClientTracer   func(c *client.Client)
}

// Record will pass along data to be sent to the tracing service (X-Ray in this case), note that this is not
// context aware. In XRayTraceStrategy, this data is sent along with the context in Capture().
func (t *XRayTraceStrategy) Record(dataType string, data interface{}) {
	// TODO: Maybe a RecordWithContext() too? Not sure if that's needed. Capture() and CaptureAsync() both will
	// have the context and at that point, Record() could simply be used again and anything needed from the context,
	// will be available.

	if t.Annotation == nil {
		t.Annotation = make(map[string]interface{})
	}
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	if t.NamespaceMetadata == nil {
		t.NamespaceMetadata = make(map[string]map[string]interface{})
	}

	if dataType != "" && data != nil {
		switch dataType {
		case "annotation", "annotations":
			for k, v := range data.(map[string]interface{}) {
				t.Annotation[k] = v
			}
		case "metadata":
			for k, v := range data.(map[string]interface{}) {
				t.Metadata[k] = v
			}
		case "namespaceMetadata", "namedMetadata", "namespace", "ns":
			for ns, metadata := range data.(map[string]map[string]interface{}) {
				for k, v := range metadata {
					if _, ok := t.NamespaceMetadata[ns]; !ok {
						t.NamespaceMetadata[ns] = make(map[string]interface{})
					}
					t.NamespaceMetadata[ns][k] = v
				}
			}
		case "error":
			t.Error = data.(error)
		}
	}
}

// Capture traces the provided synchronous function by using XRay Cpature() which puts a beginning and closing subsegment around its execution
func (t *XRayTraceStrategy) Capture(ctx context.Context, name string, fn func(context.Context) error) (err error) {
	t.setData(ctx)
	return xray.Capture(ctx, name, fn)
}

// CaptureAsync traces an arbitrary code segment within a goroutine by using XRay CaptureAsync()
func (t *XRayTraceStrategy) CaptureAsync(ctx context.Context, name string, fn func(context.Context) error) {
	t.setData(ctx)
	xray.CaptureAsync(ctx, name, fn)
}

// setData will set annotations, metadata, etc. on to the segment using context
func (t *XRayTraceStrategy) setData(ctx context.Context) {
	if t.Annotation == nil {
		for k, v := range t.Annotation {
			xray.AddAnnotation(ctx, k, v)
		}
		// Set the Annotation field nil again, do not re-apply annotations or apply them to the wrong capture
		// on subsequent calls.
		t.Annotation = nil
	}

	if t.Metadata == nil {
		for k, v := range t.Metadata {
			xray.AddMetadata(ctx, k, v)
		}
		t.Metadata = nil
	}

	if t.NamespaceMetadata == nil {
		for ns, metadata := range t.NamespaceMetadata {
			for k, v := range metadata {
				xray.AddMetadataToNamespace(ctx, ns, k, v)
			}
		}
		t.NamespaceMetadata = nil
	}

	if t.Error != nil {
		xray.AddError(ctx, t.Error)
		t.Error = nil
	}
}

// BeginSegment will begin a new trace segment, useful when running locally as AWS Lambda already does this
func (t *XRayTraceStrategy) BeginSegment(ctx context.Context, name string) (context.Context, interface{}) {
	// By default we're just using xray's BeginSegment(), but if xray is not used, then
	// something else needs to return a context and interface{} (which is ignored when running locally).
	// Note: It is possible to use xray locally still. You just need AWS credentials configured.
	// The context returned by xray.Beginsegment() is enough to use xray like normal.
	return xray.BeginSegment(ctx, name)
}

// BeginSubsegment will begin a new trace sub-segment, useful within handlers
func (t *XRayTraceStrategy) BeginSubsegment(ctx context.Context, name string) (context.Context, interface{}) {
	return xray.BeginSubsegment(ctx, name)
}

// CloseSegment will just call AWS X-Ray's Segment.Close()
func (t *XRayTraceStrategy) CloseSegment(segment interface{}, err error) {
	seg := segment.(*xray.Segment)
	seg.Close(err)
}

// CloseSubsegment will just call AWS X-Ray's Segment.Close()
func (t *XRayTraceStrategy) CloseSubsegment(segment interface{}, err error) {
	subseg := segment.(*xray.Segment)
	subseg.CloseAndStream(err)
}

// NoTraceStrategy interface does nothing (good for unit tests and disabling tracing)
type NoTraceStrategy struct {
}

// Record in this case does nothing
func (t *NoTraceStrategy) Record(dataType string, data interface{}) {}

// Capture in this case just executes the function it's wrapping
func (t *NoTraceStrategy) Capture(ctx context.Context, name string, fn func(context.Context) error) error {
	fn(ctx)
	return nil
}

// CaptureAsync in this case just executes the function it's wrapping
func (t *NoTraceStrategy) CaptureAsync(ctx context.Context, name string, fn func(context.Context) error) {
	fn(ctx)
}

// BeginSegment in this case does nothing
func (t *NoTraceStrategy) BeginSegment(ctx context.Context, name string) (context.Context, interface{}) {
	return context.Background(), nil
}

// BeginSubsegment in this case does nothing
func (t *NoTraceStrategy) BeginSubsegment(ctx context.Context, name string) (context.Context, interface{}) {
	return context.Background(), nil
}

// CloseSegment in this case does nothing
func (t *NoTraceStrategy) CloseSegment(segment interface{}, err error) {
}

// CloseSubsegment in this case does nothing
func (t *NoTraceStrategy) CloseSubsegment(segment interface{}, err error) {
}
