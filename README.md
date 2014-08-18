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


## Reccs Over MongoDB's Capped Collections

MongoDB's capped collections have a couple of distinct strict requirements
compared to Reccs. First, in MongoDB a capped collection cannot grow in data
size. This is done for performance reasons. Second, it cannot support time based
expiration. This is planned in the near future for Reccs.


## Commands Supported

Create a new collection

	reccs> CREATE foo
	OK

Delete a collection

	reccs> DELETE foo
	OK

Add an item to the collection

	reccs> ADD foo "some item"
	OK
	reccs> ADD foo "another item"
	OK

Get time ordered items in the collection 

	reccs> GET foo
	1) "some item"
	2) "another item"

Get most recent item

	reccs> HEAD foo
	"another item"

Get last item

	reccs> TAIL foo
	"some item"

Get timestamp of most recent item (returns two integers, Unix timestamp seconds
and nanoseconds)

	reccs> TSHEAD foo
	1) (integer) 1408300669
	2) (integer) 699935261

Get timestamp of last item

	reccs> TSTAIL foo
	1) (integer) 1408255676
	2) (integer) 673164618

Change the maximum number of items in the collection (default is 100)

	reccs> CSET foo maxitems 20
	OK

Ping

	reccs> PING
	PONG


## To-dos

- A command that subscribes to get new items in a collection, similar to Redis' pub/sub.
- Following the above, a "maxlisteners" config to throttle this on a per collection basis.
- Command case insensitivity.
- Time based expiration.
- Internally to process the wire protocol as a stream to handle large payloads.
- Internally better code overall. Currently this is in a "just get it working" state.


## Stability

Reccs is pretty much settled on using RESP, so all compatible redis clients and
libraries will work with reccs. Eventually commands will become set and stable
but as of right now it is subject to change, though it is unlikely.

