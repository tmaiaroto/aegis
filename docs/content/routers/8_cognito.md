# Cognito Router

There are some AWS Cognito specific events that can be routed as well. These events are for User Pool workflows.
You can find more information about the <a href="https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-identity-pools-working-with-aws-lambda-triggers.html" target="_blank">Cognito triggers here</a>.

The Cognito router is matching on the `triggerSource` value here. You also can have separate handlers for different
User Pools. You'll just need to set a `PoolID` on the router or use <span class="nowrap">`NewCognitoRouterForPool()`</span>
instead of <span class="nowrap">`NewCognitoRouter()`.</span>

Perhaps the most interesting or useful thing about this router is the ability to have your Lambda easily add
functionality to AWS Cognito. It's a great place to stick handlers to send out e-mails and more.