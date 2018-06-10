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
	"os"
)

// GetVariable will retrieve an "Aegis Variable" from Services.Variables or OS environment variable fallback
func (a *Aegis) GetVariable(key string) string {
	valueStr := ""

	if val, ok := a.Services.Variables[key]; ok {
		valueStr = val
	}

	// Still empty? Try plain environment variable.
	if valueStr == "" {
		valueStr = os.Getenv(key)
	}
	return valueStr
}
