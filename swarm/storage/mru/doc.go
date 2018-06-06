// Mutable resource is an entity which allows updates to a resource
// without resorting to ENS on each update.
// The update scheme is built on swarm chunks with chunk keys following
// a predictable, versionable pattern.
//
// Updates are defined to be periodic in nature, where periods are
// expressed in terms of number of blocks.
//
// The root entry of a mutable resource is tied to a unique identifier,
// typically - but not necessarily - an ens name.  The identifier must be
// an valid IDNA string. It also contains the block number
// when the resource update was first registered, and
// the block frequency with which the resource will be updated, both of
// which are stored as little-endian uint64 values in the database (for a
// total of 16 bytes). It also contains the unique identifier.
// This MRU info is stored in a separate content-addressed chunk
// (call it the metadata chunk), with the following layout:
//
// (startblock|frequency|address|identifier)
//
// (The two first zero-value bytes are used for disambiguation by the chunk validator,
// and update chunk will always have a value > 0 there.)
//
// The root entry tells the requester from when the mutable resource was
// first added (block number) and in which block number to look for the
// actual updates. Thus, a resource update for identifier "føø.bar"
// starting at block 4200 with frequency 42 will have updates on block 4242,
// 4284, 4326 and so on.
//
// Actual data updates are also made in the form of swarm chunks. The keys
// of the updates are the hash of a concatenation of properties as follows:
//
// sha256(period|version|address|namehash)
//
// The period is (currentblock - startblock) / frequency
//
// Using our previous example, this means that a period 3 will have 4326 as
// the block number.
//
// If more than one update is made to the same block number, incremental
// version numbers are used successively.
//
// A lookup agent need only know the identifier name in order to get the versions
//
// the resourcedata is:
// headerlength|period|version|identifier|data
//
// the full update data that goes in the chunk payload is:
// resourcedata|sign(resourcedata)
//
// headerlength is a 16 bit value containing the byte length of period|version|name
package mru
