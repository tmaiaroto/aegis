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

package main

import (
	"encoding/json"
	"github.com/tmaiaroto/aegis/lambda"
	//"github.com/labstack/echo"
	//"log"
	//"net/http"
)

func main() {
	lambda.HandleProxy(func(context *lambda.Context, eventJSON json.RawMessage) (interface{}, error) {

		var v map[string]interface{}
		if err := json.Unmarshal(eventJSON, &v); err != nil {
			return nil, err
		}
		//log.Println(v)
		vStr, _ := json.Marshal(v)
		return string(vStr), nil
		// return v, nil

	})

	// e := echo.New()
	// e.GET("/", func(c echo.Context) error {

	// 	return c.JSON(http.StatusCreated, map[string]string{"foo": "bar"})
	// })

	// if err := e.Start(":9500"); err != nil {
	// 	e.Logger.Fatal(err.Error())
	// }
}
