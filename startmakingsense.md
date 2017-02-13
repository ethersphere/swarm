# META PROTOCOL - DRAFT IMPLEMENTATION

## TYPES

### QUERY

0. MATCH_REQUEST (Does this exist, I don't have a clue) 
1. MATCH_RESPONSE (I say this is what you're looking for)
2. MATCH_CONFLICT (you're lying...?)
3.

### ANNOUNCEMENT

0. Custom
1. Work (atomic unit)
2. Recording (persistent rendering)
3. Performance (ephemeral rendering)

### PAYLOAD

0. Custom
1. Authiddata 
2. Workdata
3. Artistdata
4. Mediadata
5. Licensedata  
6. Usagedata
7. 
8.
9.
10.
11.
12.
13.
14.
15.


## META HEADERS

### META QUERY HEADER

|field|bits|content|
|-|-|-|
|protocolversion|0-2|1-4|
|QUERY_TYPE|3-6| 4 slots, see "types"|
|crc|7-166|SHA-1 hash of the payload of this message, used as matching ID for response|
|unused|167|...|
|payload|168-|data|


### META ANNOUNCE HEADER

|field|bits|content|
|-|-|-|
|protocolversion|0-2|1-4|
|ANNOUNCEMENT_TYPE|3-6| 4 slots, see "types"|
|timestamp|7-70|64 bit milliseconds|
|payload|71-|data|

## PAYLOAD INTERPRETATION EXAMPLES

### QUERY A ARTIST/WORKNAME PAIR

|META.LOOKUP|content|bits|part|
|-|-|-|-|
|protocolversion|1|0-2|HEAD|
|querytype|1 (MATCH_REQUEST)|3-6|HEAD|
|crc|#sha1 of payload#|7-166|HEAD|
|unused|0|167|HEAD|
|payload type (1)|3 (ARTISTDATA)|168-172|BODY / PAYLOAD 1|
|unused|0|173|BODY / PAYLOAD 1|
|itemcount|1|174-176|BODY / PAYLOAD 1|
|labellength|4|177-182|BODY / PAYLOAD 1|
|datalength|11|183-192|BODY / PAYLOAD 1|
|label|name|193-196|BODY / PAYLOAD 1|
|data|the police|197-207|BODY / PAYLOAD 1|
|payload type (2)|2 (WORKDATA)|208-212|BODY / PAYLOAD 2|
|unused|0|213|BODY / PAYLOAD 2|
|itemcount|1|214-216|BODY / PAYLOAD 2|
|labellength|4|217-222|BODY / PAYLOAD 2|
|datalength|12|223-232|BODY / PAYLOAD 2|
|label|name|233-236|BODY / PAYLOAD 2|
|data|synchronicity|237-248|BODY / PAYLOAD 2|

### RETURN WORK

(bit counts are off here)

|META.LOOKUP|content|bits|part|
|-|-|-|-|
|protocolversion|1|0-2|HEAD|
|querytype|2 (MATCH_RESPONSE)|3-6|HEAD|
|crc|#crc from request#|7-26|HEAD|
|unused|0|27-31|HEAD|
|payload type (1)|4 (MEDIADATA)|73-80|BODY / PAYLOAD 1|
|unused|0|81|BODY / PAYLOAD 1|
|itemcount|1|82-87|BODY / PAYLOAD 1|
|labellength|0|88-93|BODY / PAYLOAD 1|
|datalength|0|94-107|BODY / PAYLOAD 1|
|data|#swarmhash#|108-139|BODY / PAYLOAD 1|
|payload type (2)|5 (LICENCEDATA)|140-147|BODY / PAYLOAD 2|
|unused|0|148|BODY / PAYLOAD 2|
|itemcount|1|149-151|BODY / PAYLOAD 2|
|labellength|0|152-157|BODY / PAYLOAD 2|
|datalength|0|158-167|BODY / PAYLOAD 2|
|data|#swarmhash#|168-199|BODY / PAYLOAD 2|
|payload type(3)|1 (AUTHIDDATA)|200-207|BODY / PAYLOAD 3|
|unused|0|208|BODY / PAYLOAD 3|
|itemcount|1|209-211|BODY / PAYLOAD 3|
|labellength|4|212-217|BODY / PAYLOAD 3|
|datalength|15|218-227|BODY / PAYLOAD 3|
|label|iswc|228-231|BODY / PAYLOAD 3|
|data|T-010.041.364-6|232-246|BODY / PAYLOAD 3|
