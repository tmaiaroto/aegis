package apigateway

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewSwagger(t *testing.T) {
	testCfg := SwaggerConfig{
		Title:     "Test Swagger API",
		LambdaURI: "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegistest/invocation",
	}

	testCfgWithoutTitle := SwaggerConfig{
		LambdaURI: "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegistest/invocation",
	}
	Convey("A new API Swagger struct should be returned", t, func() {
		apiSwagger, _ := NewSwagger(&testCfg)
		So(apiSwagger, ShouldHaveSameTypeAs, Swagger{})
		So(apiSwagger.Info.Title, ShouldEqual, testCfg.Title)
		So(apiSwagger.Paths, ShouldContainKey, "/")
		So(apiSwagger.Paths, ShouldContainKey, "/{proxy+}")
		So(apiSwagger.Paths["/{proxy+}"].XAmazonAPIGatwayAnyMethod.XAmazonAPIGatewayIntegration.CacheNamespace, ShouldHaveLength, 6)

		apiSwaggerDefaultTitle, _ := NewSwagger(&testCfgWithoutTitle)
		So(apiSwaggerDefaultTitle.Info.Title, ShouldEqual, "Example Aegis API")
	})

	Convey("A new API Swagger must be given a LambdaURI", t, func() {
		_, err := NewSwagger(&SwaggerConfig{})
		So(err, ShouldNotBeNil)
	})
}

func TestGetLambdaURI(t *testing.T) {
	testLambdaARN := "arn:aws:lambda:us-east-1:12345:function:aegis_example:6"
	expectedLambdaURI := "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegis_example:6/invocations"

	testVersionlessLambdaARN := "arn:aws:lambda:us-east-1:12345:function:aegis_example"
	expectedVersionlessLambdaURI := "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:12345:function:aegis_example/invocations"

	Convey("A Lambda URI should be formatted and returned based on a given Lambda ARN", t, func() {
		So(GetLambdaURI(testLambdaARN), ShouldEqual, expectedLambdaURI)
		So(GetLambdaURI(testVersionlessLambdaARN), ShouldEqual, expectedVersionlessLambdaURI)
	})
}

func TestRandomCacheNamespace(t *testing.T) {
	Convey("A random character string should be returned with given length", t, func() {
		So(randomCacheNamespace(6), ShouldHaveLength, 6)
		So(randomCacheNamespace(6), ShouldHaveSameTypeAs, "")
		So(randomCacheNamespace(100), ShouldHaveLength, 100)
	})
}
