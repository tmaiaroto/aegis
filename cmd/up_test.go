package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestUpCmd(t *testing.T) {
	Convey("compress", t, func() {
		Convey("Should return the path to a Lambda function zip file", func() {
			testZipFileName := "./aegis_function.test.zip"
			filePath := compress(testZipFileName)
			// make the "aegis_app" file so it exists to zip
			_, _ = os.Create(aegisAppName)
			So(filePath, ShouldContainSubstring, "aegis_function.test.zip")

			// cleanup
			_ = os.Remove(aegisAppName)
			_ = os.Remove(testZipFileName)
		})
	})

	Convey("getAccountInfoFromLambdaArn", t, func() {
		Convey("Should return account info from a given Lamba ARN", func() {
			arn := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example:1"
			account, region := getAccountInfoFromLambdaArn(arn)
			So(account, ShouldEqual, "1234567890")
			So(region, ShouldEqual, "us-east-1")
		})
	})

	Convey("stripLamdaVersionFromArn", t, func() {
		Convey("Should strip the Lambda version number from a given Lambda ARN", func() {
			arn := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example:1"
			expected := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example"
			So(stripLamdaVersionFromArn(arn), ShouldEqual, expected)
		})

		Convey("Should return the versionless ARN if no version was given", func() {
			arn := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example"
			expected := "arn:aws:lambda:us-east-1:1234567890:function:aegis_example"
			So(stripLamdaVersionFromArn(arn), ShouldEqual, expected)
		})
	})

	Convey("getExecPath", t, func() {
		Convey("Should return a given executable file's path", func() {
			So(getExecPath("go"), ShouldNotBeEmpty)
		})
	})
}
