## Retreive node information

`pss_getPublicKey`

parameters:
none

returns:
1. publickey `base64(bytes)` `[]byte`

`pss_baseAddr`

parameters:
none

returns:
1. swarm overlay address `base64(bytes)` `[]byte`

`pss_stringToTopic`

parameters:
1. topic string `string` `string`

returns:
1. pss topic `base64(bytes[4])` `pss.Topic`

# Receive messages

`pss_subscribe`*

parameters:

1. "receive" (literal) `string` `string` 
2. topic `base64(bytes)` `pss.Topic`

returns:

1. subscription handle `base64(byte)` `rpc.ClientSubscription`

* In `golang` as special method is used:

`rpc.Client.Subscribe(context.Context, "pss", chan pss.APIMsg, "receive", pss.Topic)`

Incoming messages are encapsulated in an object (`pss.APIMsg` in `golang`) with the following members:

1. Msg (message payload) `base64(bytes)` `[]byte`
2. Asymmetric (true if encrypted with public key cryptography) `bool` `bool`
3. Key (raw encryption key in hex format) `string` `string`

## Send message using asymmetric encryption

`pss_setPeerPublicKey`

parameters:

1. public key of peer `base64(bytes)` `[]byte`
2. topic `base64(bytes)` `pss.Topic`
3. address of peer `base64(bytes)` `pss.PssAddress`

returns:

none

`pss_sendAsym`

parameters:

1. public key of peer `base64(bytes)` `[]byte`
2. topic `base64(bytes)` `pss.Topic`
3. message `base64(bytes)` `[]byte`

returns:

none

## Send message using symmetric encryption

`pss_setSymmetricKey`

parameters:

1. symmetric key `base64(bytes)` `[]byte`
2. topic `base64(bytes)` `pss.Topic`
3. address of peer `base64(bytes)` `pss.PssAddress`
4. use for decryption `bool` `bool`

returns:

1. symmetric key id `string` `string`

`pss_sendSym`

parameters:

1. symmetric key id `string` `string`
2. topic `base64(bytes)` `pss.Topic`
3. message `base64(bytes)` `[]byte`

returns:

none

## Querying peer keys

`pss_GetSymmetricAddressHint`

parameters:

1. topic `base64(bytes)` `pss.Topic`
3. symmetric key id `string` `string`

returns:

1. associated address of peer `base64(bytes)` `pss.PssAddress`

`pss_GetAsymmetricAddressHint`

parameters:

1. topic `base64(bytes)` `pss.Topic`
3. public key in hex form `string` `string`

returns:

1. associated address of peer `base64(bytes)` `pss.PssAddress`

## Handshake module

`pss_addHandshake`

parameters:

1. topic to activate handshake on `base64(bytes)` `pss.Topic`

returns:

none

`pss_removeHandshake`

parameters:

1. topic to remove handshake from `base64(bytes)` `pss.Topic`

returns:

none

`pss_handshake`

parameters:

1. public key of peer in hex format `string` `string`
2. topic `base64(bytes)` `pss.Topic`
3. block calls until keys are received `bool` `bool`
4. flush existing incoming keys `bool` `bool`

returns:

1. list of symmetric keys `string[]` `[]string`*

* If parameter 3 is false, the returned array will be empty.

`pss_getHandshakeKeys`

parameters:

1. public key of peer in hex format `string` `string`
2. topic `base64(bytes)` `pss.Topic`
3. include keys for incoming messages `bool` `bool`
4. include keys for outgoing messages `bool` `bool`

returns:

1. list of symmetric keys `string[]` `[]string`

`pss_getHandshakeKeyCapacity`

parameters:

1. symmetric key id `string` `string`

return:

1. remaining number of messages key is valid for `uint` `uint16`

`pss_getHandshakePublicKey`

parameters:

1. symmetric key id `string` `string`

returns:

1. Associated public key in hex format `string` `string`

`pss_releaseHandshakeKey`

parameters:

1. public key of peer in hex format `string` `string`
2. topic `base64(bytes)` `pss.Topic`
3. symmetric key id to release `string` `string`
4. remove keys instantly `bool` `bool`

returns:

1. whether key was successfully removed `bool` `bool`
