# LRPC
AWS Lambda net/rpc server, client via AWS API.

The packages provided are intended to allow deploying
net/rpc services to AWS Lambda and calling them with
net/rpc Client conventions or via a JSON-RPC 1.0 payload.

## Server
Servers use Register or RegisterName from lrpc/server, or
RPCServer to set the net/rpc server to serve. The
server.Serve() function is then called from a main package.
Deploying such main packages is facilitated with 
[lago](https://github.com/cloudinterfaces/lago).

## Client
Clients access net/rpc servers running as Lambda handlers
via the lrpc/client package. Call (and Invoke) issues
the request via the AWS Lambda API.

## Demo
See [the demo package](demo).

## How it works
The server inspects the payload and handles the request
as gob-encoded or a JSON-RPC 1.0

### With client package
The client gob-encodes an rpc.Request and request body
to the Invoke payload. The server invokes the rpc.Request
with the supplied body. If a method error occurs, it is 
encoded as a rpc.Response. Otherwise, the response body
is also encoded.

### With JSON-RPC
The JSON-RPC request is sent as a Lambda payload directly,
either via REST call or any AWS SDK. The "id" field of the 
request may be of any type. If the "id" is null or 
omitted, the output of any method call is discarded. Otherwise,
the "id" field as supplied will be returned with the response.
Method call output is marshalled to the "result" field. If
an error occurs or a panic is recovered at any point
in the Lambda invocation, the "error" field will be populated
with a message and the "result" field will be null. All marshalling
and unmarshalling is by the json package's conventions.

## Notes
The Lambda environment must be considered stateless
(unless you're aware of how it isn't and design your
net/rpc service accordingly). If a server cannot decode
a request (which should be impossible), an out-of-band error occurs.
In-band errors use a unique convention: errors returned
by a method invocation append a newline followed by
the request ID, errors and panics that occur before
or after a method invocation append a tab character followed
by the request ID. The intent is not nessisarily for
clients to inspect returned delimiters but provide
additional information in a unique way when read by a human. Errors
returned by the client (in general due to invalid invocation)
are unadorned.

## Related
The [lh](https://github.com/cloudinterfaces/lh) package makes it easy to serve (many or most) http.Handlers with the AWS Lambda Go runtime.
The [lago](https://github.com/cloudinterfaces/lago) tool makes it easy to deploy Go
handlers to the Lambda Go runtime.
