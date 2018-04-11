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

package framework

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// TraceStrategy interface uses XRay by default. Any struct implementing this interface can be used instead.
type TraceStrategy struct {
	Annotations     map[string]interface{}
	Metadata        map[string]interface{}
	AWSClientTracer func(c *client.Client)
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
