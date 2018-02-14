# LRPC demo

A Lambda function configured with the Go runtime must exist (it may be created via the AWS web interface or by other means). With lrpc package in your GOPATH and sufficient AWS credentials in your home directory the server may be deployed with:
```lago deploy -func fname -target github.com/cloudinterfaces/lrpc/demo/server```
substituting fname above with whatever the function name you wish to deploy to is and assuming it exists in the region us-east-1.

To test it demo/client will make several calls:
```go run $GOPATH/src/github.com/cloudinterfaces/lrpc/demo/client/clientdemo.go fname```
Again subsituting fname with whatever the function name is.

The JSON-RPC functionality can also be tested with a test event configured in the AWS Lambda user interface. For example:
```{"method":"Arith.Divide","params":{"A":5,"B":2},"id":"one"}```
