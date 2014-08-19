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
compared to Reccs. This is mainly to optimize MongoDB performance.

You cannot do the following in MongoDB capped collections:

- grow the document capacity in a capped collection
- combine a document capacity with a TTL (time based expiration)

With reccs, you can.


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

Get time ordered items in the collection. This includes the timestamps (seconds,
then nanoseconds).

	reccs> GET foo
	1) 1) (integer) 1408255676
	   2) (integer) 673164618
	   3) "some item"
	2) 1) (integer) 1408300669
	   2) (integer) 699935261
	   3) "another item"

Get most recent item

	reccs> HEAD foo
	1) (integer) 1408300669
	2) (integer) 699935261
	3) "another item"

Get last item

	reccs> TAIL foo
	1) (integer) 1408255676
	2) (integer) 673164618
	3) "some item"

To get just the timestamps in these commands, append with a "T"

	reccs> GETT foo
	1) 1) (integer) 1408255676
	   2) (integer) 673164618
	2) 1) (integer) 1408300669
	   2) (integer) 699935261

	reccs> HEADT foo
	1) (integer) 1408300669
	2) (integer) 699935261

	reccs> TAILT foo
	1) (integer) 1408255676
	2) (integer) 673164618

To get just the data in these commands, append with a "D"
Get timestamp of last item

	reccs> GETD foo
	"some item"
	"another item"

	reccs> HEADD foo
	"another item"

	reccs> TAILD foo
	"some item"

Change the maximum number of items in the collection (default is 100)

	reccs> CSET foo maxitems 20
	OK

Change the max age of items in the collection in milliseconds (default is no 
limit set to '0')

	reccs> CSET foo maxage 60000
	OK

Ping

	reccs> PING
	PONG


## To-dos

- A command that subscribes to get new items in a collection, similar to Redis' pub/sub.
- Following the above, a "maxlisteners" config to throttle this on a per collection basis.
- Internally to process the wire protocol as a stream to handle large payloads.


## Stability

Reccs is pretty much settled on using RESP, so all compatible redis clients and
libraries will work with reccs. Eventually commands will become set and stable
but as of right now it is subject to change, though it is unlikely.

