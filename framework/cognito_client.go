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
	"bytes"
	"context"
	"crypto/rsa"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

// CognitoAppClient is an interface for working with AWS Cognito
type CognitoAppClient struct {
	Region                   string
	UserPoolID               string
	ClientID                 string
	UserPoolType             *cognitoidentityprovider.UserPoolType
	UserPoolClient           *cognitoidentityprovider.UserPoolClientType
	WellKnownJWKs            *jwk.Set
	BaseURL                  string
	HostedLoginURL           string
	HostedLogoutURL          string
	HostedSignUpURL          string
	RedirectURI              string
	LogoutRedirectURI        string
	TokenEndpoint            string
	Base64BasicAuthorization string
	Tracer                   XRayTraceStrategy
}

// CognitoAppClientConfig defines required info to build a new CognitoAppClient
type CognitoAppClientConfig struct {
	Region            string                 `json:"region"`
	PoolID            string                 `json:"poolId"`
	ClientID          string                 `json:"clientId"`
	RedirectURI       string                 `json:"redirectUri"`
	LogoutRedirectURI string                 `json:"logoutRedirectUri"`
	TraceContext      context.Context        `json:"-"`
	AWSClientTracer   func(c *client.Client) `json:"-"`
}

// CognitoToken defines a token struct for JSON responses from Cognito TOKEN endpoint
type CognitoToken struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// NewCognitoAppClient returns a new CognitoAppClient interface configured for the given Cognito user pool ID and client ID
// NOTE: Best to not call this on every event handle because it makes 3 HTTP requests for data, capitalize on container re-use.
func NewCognitoAppClient(cfg *CognitoAppClientConfig) (*CognitoAppClient, error) {
	var err error
	c := &CognitoAppClient{
		Region:            cfg.Region,
		UserPoolID:        cfg.PoolID,
		ClientID:          cfg.ClientID,
		RedirectURI:       cfg.RedirectURI,
		LogoutRedirectURI: cfg.LogoutRedirectURI,
	}

	// Set the PoolType, it contains a bunch of useful info
	sess, err := session.NewSession()
	if err != nil {
		log.Println("could not get Cognito UserPoolType, session could not be created")
		return c, err
	}
	svc := cognitoidentityprovider.New(sess)
	// Wrap in XRay so it gets logged and appears in service map
	cfg.AWSClientTracer(svc.Client)

	// userPoolOutput, err := svc.DescribeUserPool(&cognitoidentityprovider.DescribeUserPoolInput{
	userPoolOutput, err := svc.DescribeUserPoolWithContext(cfg.TraceContext, &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(c.UserPoolID),
	})
	if err == nil {
		c.UserPoolType = userPoolOutput.UserPool
	} else {
		log.Println("Error getting Cognito user pool", err)
	}

	// Get the client app
	// TODO: If cfg.TraceContext == nil, use this normal method to get the client. This would mean no tracing.
	// userPoolClientOutput, err := svc.DescribeUserPoolClient(&cognitoidentityprovider.DescribeUserPoolClientInput{
	userPoolClientOutput, err := svc.DescribeUserPoolClientWithContext(cfg.TraceContext, &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(c.ClientID),
		UserPoolId: aws.String(c.UserPoolID),
	})
	if err == nil {
		c.UserPoolClient = userPoolClientOutput.UserPoolClient
	} else {
		log.Println("Error getting Cognito user pool client", err)
	}
	if c.UserPoolClient != nil {
		// Set the Base64 <client_id>:<client_secret> for basic authorization header
		var buffer bytes.Buffer
		buffer.WriteString(c.ClientID)
		buffer.WriteString(":")
		buffer.WriteString(aws.StringValue(c.UserPoolClient.ClientSecret))
		base64AuthStr := b64.StdEncoding.EncodeToString(buffer.Bytes())
		buffer.Reset()

		buffer.WriteString("Basic ")
		buffer.WriteString(base64AuthStr)
		c.Base64BasicAuthorization = buffer.String()
		buffer.Reset()

		// Set up login and signup URLs, if there is a domain available
		c.getURLs()
	}

	// Set the well known JSON web token key sets
	err = c.getWellKnownJWTKs()
	if err != nil {
		log.Println("Error getting well known JWTKs", err)
	}

	return c, err
}

// getWellKnownJWTKs gets the well known JSON web token key set for this client's user pool
func (c *CognitoAppClient) getWellKnownJWTKs() error {
	// https://cognito-idp.<region>.amazonaws.com/<pool_id>/.well-known/jwks.json
	var buffer bytes.Buffer
	buffer.WriteString("https://cognito-idp.")
	buffer.WriteString(c.Region)
	buffer.WriteString(".amazonaws.com/")
	buffer.WriteString(c.UserPoolID)
	buffer.WriteString("/.well-known/jwks.json")
	wkjwksURL := buffer.String()
	buffer.Reset()

	// Use this cool package
	set, err := jwk.Fetch(wkjwksURL)
	if err == nil {
		c.WellKnownJWKs = set
	} else {
		log.Println("There was a problem getting the well known JSON web token key set")
		log.Println(err)
	}
	return err
}

// getURLs gets all of the URLs and endpoints for the Cognito client, AWS hosted login/signup pages, token endpoints for oauth2, etc.
func (c *CognitoAppClient) getURLs() {
	if c.UserPoolType != nil && c.UserPoolType.Domain != nil {
		// Get the base URL
		var buffer bytes.Buffer
		buffer.WriteString("https://")
		buffer.WriteString(aws.StringValue(c.UserPoolType.Domain))
		buffer.WriteString(".auth.")
		buffer.WriteString(c.Region)
		buffer.WriteString(".amazoncognito.com")
		baseURL := buffer.String()
		c.BaseURL = baseURL
		buffer.Reset()

		// Set the HostedLoginURL
		buffer.WriteString(baseURL)
		buffer.WriteString("/login?response_type=code&client_id=")
		buffer.WriteString(c.ClientID)
		buffer.WriteString("&redirect_uri=")
		buffer.WriteString(c.RedirectURI)
		c.HostedLoginURL = buffer.String()
		buffer.Reset()

		// Set the HostedLogoutURL
		buffer.WriteString(baseURL)
		buffer.WriteString("/logout?response_type=code&client_id=")
		buffer.WriteString(c.ClientID)
		buffer.WriteString("&redirect_uri=")
		buffer.WriteString(c.RedirectURI)
		c.HostedLogoutURL = buffer.String()
		buffer.Reset()

		// Set the HostedSignUpURL
		buffer.WriteString(baseURL)
		buffer.WriteString("/signup?response_type=code&client_id=")
		buffer.WriteString(c.ClientID)
		buffer.WriteString("&redirect_uri=")
		buffer.WriteString(c.RedirectURI)
		c.HostedSignUpURL = buffer.String()
		buffer.Reset()

		// Set the authorization token URL
		buffer.WriteString(c.BaseURL)
		buffer.WriteString("/oauth2/token")
		c.TokenEndpoint = buffer.String()
		buffer.Reset()
	}
}

// GetTokens will make a POST request to the Cognito TOKEN endpoint to exchange a code for an access token
func (c *CognitoAppClient) GetTokens(code string, scope []string) (CognitoToken, error) {
	var token CognitoToken

	hc := http.Client{}
	// set the url-encoded payload
	form := url.Values{}
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.ClientID)
	form.Set("redirect_uri", c.RedirectURI)
	if len(scope) > 0 {
		form.Set("scope", strings.Join(scope, " "))
	}
	// request
	req, err := http.NewRequest("POST", c.TokenEndpoint, strings.NewReader(form.Encode()))
	if err == nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		// This should be a string like: Basic XXXXXXXXXX
		req.Header.Add("Authorization", c.Base64BasicAuthorization)

		resp, err := hc.Do(req)
		if err != nil {
			log.Println("Could not make request to Cognito TOKEN endpoint")
			return token, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Could not read response body from Cognito TOKEN endpoint")
			return token, err
		}

		err = json.Unmarshal(body, &token)
		if err != nil {
			log.Println("Could not unmarshal token response from Cognito TOKEN endpoint")
		}
	} else {
		log.Println("Error making HTTP request", err)
	}
	return token, err
}

// ParseAndVerifyJWT will parse and verify a JWT, if an error is returned the token is invalid,
// only a valid token will be returned
//
// https://github.com/awslabs/aws-support-tools/tree/master/Cognito/decode-verify-jwt
// Amazon Cognito returns three tokens: the ID token, access token, and refresh token—the ID token
// contains the user fields defined in the Amazon Cognito user pool.
//
// To verify the signature of an Amazon Cognito JWT, search for the key with a key ID that matches
// the key ID of the JWT, then use libraries to decode the token and verify the signature.
//
// Be sure to also verify that:
//  - The token is not expired.
//  - The audience ("aud") in the payload matches the app client ID created in the Cognito user pool.
func (c *CognitoAppClient) ParseAndVerifyJWT(t string) (*jwt.Token, error) {
	// 3 tokens are returned from the Cognito TOKEN endpoint; "id_token" "access_token" and "refresh_token"
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		// Looking up the key id will return an array of just one key
		keys := c.WellKnownJWKs.LookupKeyID(token.Header["kid"].(string))
		if len(keys) == 0 {
			log.Println("Failed to look up JWKs")
			return nil, errors.New("could not find matching `kid` in well known tokens")
		}
		// Build the public RSA key
		key, err := keys[0].Materialize()
		if err != nil {
			log.Printf("Failed to create public key: %s", err)
			return nil, err
		}
		rsaPublicKey := key.(*rsa.PublicKey)
		return rsaPublicKey, nil
	})

	// Populated when you Parse/Verify a token
	// First verify the token itself is a valid format
	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Then check time based claims; exp, iat, nbf
			err = claims.Valid()
			if err == nil {
				// Then check that `aud` matches the app client id
				// (if `aud` even exists on the token, second arg is a "required" option)
				if claims.VerifyAudience(c.ClientID, false) {
					return token, nil
				} else {
					err = errors.New("token audience does not match client id")
					log.Println("Invalid audience for id token")
				}
			} else {
				log.Println("Invalid claims for id token")
				log.Println(err)
			}
		}
	} else {
		log.Println("Invalid token:", err)
	}

	return nil, err
}
