# Reccs

Reccs is a REmote Capped Collection Server providing functionality similar to 
MongoDB's capped collection, communicating with the Redis Serialization 
Protocol, RESP. You can interact with reccs by using the redis cli or any redis
library in your programming language.

A capped collection is a first in, first out store that is constraint to a
certain capacity. This is ideal for time based or ordered data where only a
certain time or frame is needed.

Reccs is a disk based store, relying on the filesystem and OS to provide good
enough performance and caching for reads and writes. 


## Commands Supported

Create a new collection

	reccs> CREATE foo
	OK

Delete a collection

	reccs> DELETE foo
	OK

Add an item to the collection

	reccs> ADD foo someitem
	OK
	reccs> ADD foo anotheritem
	OK

Get time ordered items in the collection 

	reccs> GET foo
	1) "someitem"
	2) "anotheritem"

Get most recent item

	reccs> HEAD foo
	"anotheritem"

Get last item

	reccs> TAIL foo
	"someitem"

Get timestamp of most recent item (returns two integers, Unix timestamp seconds
and nanoseconds)

	reccs> TSHEAD foo
	1) (integer) 1408300669
	2) (integer) 699935261

Get timestamp of last item

	reccs> TSTAIL foo
	1) (integer) 1408255676
	2) (integer) 673164618

Ping

	reccs> PING
	PONG


## Stability

The Redis protocol support does not support pipelining and pub/sub yet. Adding
the support is still to be determined. 
