/*
Package mru defines Mutable resource updates.

A Mutable Resource is an entity which allows updates to a resource
without resorting to ENS on each update.
The update scheme is built on swarm chunks with chunk keys following
a predictable, versionable pattern.

Updates are defined to be periodic in nature, where the update frequency
is expressed in seconds.

A Resource is tied to a unique identifier that is deterministically generated out of
the metadata content that describes it. This metadata includes a user-defined topic, a resource
start time that indicates when the resource becomes valid, the frequency in seconds with
which the resource is expected to be updated.

A Resource View is defined as a specific user's point of view about a particular resource.
Thus, a View is a Resource + the user's address (userAddr)

The Resource structure tells the requester from when the mutable resource was
first added (Unix time in seconds) and in which moments to look for the
actual updates. Thus, a Resource with Topic "føø.bar"
starting at unix time 1528800000 with frequency 300 (every 5 mins) will have updates on 1528800300,
1528800600, 1528800900 and so on.

Actual data updates are also made in the form of swarm chunks. The keys
of the updates are the hash of a concatenation of properties as follows:

updateAddr = H(View, period, version)
where H is the SHA3 hash function
The period is (currentTime - startTime) / frequency

Using our previous example, this means that a period 3 will happen when the
clock hits 1528800900

If more than one update is made in the same period, incremental
version numbers are used successively.

A user looking up a resource would only need to know the View in order to
another user's updates

the resource update data is:
resourcedata = View|period|version|data

the full update data that goes in the chunk payload is:
resourcedata|sign(resourcedata)
*/
package mru
