# P2P SIMULATIONS

This functionality is very rudimentary, and specifics may change without notice.

## HTTP SERVER

Provide entrypoints for HTTP API calls. 

Generic controllers are found in `./session_controller.go`

The HTTP muxer (`./rest_api_server.go`)analyses the URI sent in the HTTP request. The `NewSessionController()` returns a controller that can resolve the analyzed request:

* `ResourceController.Resource()` finds `ResourceController` to match the endpoint.
* `ResourceController.Handler()` finds `ResourceHandler`to match the HTTP method.

## CONTROLLERS

The `ResourceHandlers` struct of a `ResourceController` should implement one or more *CRUD* requests, with symbols

| ResourceHandler Symbol | ..maps to.. |HTTP   |
|-----------------|-------------|-------|
| Create          |      ->     |POST   |
| Retrieve	      |      ->     |GET    |
| Update          |      ->     |PUT    |
| Delete          |      ->     |DELETE |

Each of these symbols should point to a `ResourceHandler` that implements:

* `Handler` - the function actually performing the action on the entrypoint
* `Type` - optional reflection of the data type expected in the `msg interface{}` input param. **behavior is undefined if you do not include this** The input param struct will be populated by the body of the HTTP request in JSON format. 

All Controllers should implement interface `Controller` defined in `rest_api_server.go`


## GENERIC ENTRYPOINTS

*(Destroy handlers are deliberately not mentioned here)*

The `ResourceController` residing directly below the SessionController is available at entrypoint "/" in HTTP requests.


In the generic architecture in `session_controller.go` the `NewNetworkController()` call implements an entrypoint accessible by the networkname passed to it through the NetworkConfig input struct. Thus:

```
HTTP POST / HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 15

{"Id":"foobar"}
```

will create an endpoint `/foobar` in which the handlers of the "NetworkController" can be reached.

### NETWORK CONTROLLER


The "NetworkController" currently only implements an graph describing events since last equivalent call. This graph can in turned be used by a layer the monitors, visualizes, reports etc...

### NODE CONTROLLER

Resides on the `/[networkname]/node` entrypoint, and currently implements:

* **Create a node**: HTTP POST without body
* **Return a list of nodes**: HTTP GET without body
* **Toggle node up / down**: HTTP PUT with 1D hash: `{"One": [int:indexofnodebycreationsequence]}`
* **Toggle connection of nodes**: HTTP PUT with 1D hash: `{"One": [uint:indexofnodebycreationsequence], "Other": [uint:indexofnodebycreationsequence]}`

To send message between connected nodes, intention is something like **HTTP PUT** with: `{"One": [uint:indexofnodebycreationsequence], "Other": [uint:indexofnodebycreationsequence], "AssetType": [uint:protocols.Codemap], "Data": {[data struct as defined by codemap]}}`. This must for the time being be implemented manually, but an illustrative example is something like:

`net.Send(onenode.Id, othernode.Id, uint64(net.Ct.GetCode(protomsg)), protomsg)`

(sorry, the documentation of the *protocols* layer is another story)

### DEBUG CONTROLLER

Dumps in log form all events in the session. It is **not** intended for use in actual simulation implementations, because it will most likely flush the event buffer.


## JOURNAL

*Currently it's not safe to assume that events are hooked in the p2p layer, they may only be registered in the simulation layer*

Certain phenomena the p2p stack generate events. The simulation framework provides a `Journal` which subscribes to an event muxer passed from the "SessionController" and up the stack. It is this `Journal` that feeds the event graph in the (current) NetworkController output.

The `Journal` has a cursor which can be reset, and the `Journal` can thus be replayed, making it possible to store and repeat sessions. There is currently no generic endpoint implemented for this.

## INVOCATION

Feneric functionality can be invoked by running `go run -v ./connectivity.go` in `./examples` and interfacing `http://127.0.0.1:8888`.

The general idea is that whatever layer that should make use of the events emanating from the network simulation retrieve these by periodically calling `GET /[networkname]/ HTTP/1.0` (see `session_controller_test.go:TestUpdate()` for an example of what this output looks like).

The events in the network simulation can originate from backend code or also by using the HTTP interface. `session_controller_test.go` also illustrates some examples of the latter.
