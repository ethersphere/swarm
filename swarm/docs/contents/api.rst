*********************
API
*********************

Command-line API
============================

The command-line parameters to @command{geth} affecting swarm at present
are:

@deffn {} bzzaccount (coinbase)
Ethereum account address serving as the swarm node base address (routing key and storage centroid)
@end deffn

@deffn {} bzzconfig (@file{<datadir>/bzz/<bzzaccount>/config.json})
swarm configuration file (json)
@end deffn

Console/IPC API
=======================

@deffn {} @code{bzz.get(path)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.put(content, content_type)}
@deffnx {content} {}
@deffnx {content_type} {}
@end deffn

@deffn {} @code{bzz.modify(root_hash, path, content_hash, content_type)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.download(path, local_path)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.upload(local_path, index)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.register(sender, domain, hash)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.resolve(host)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.info()}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.issue(beneficiary, amount)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.cash(cheque)}
@deffnx {} {}
@end deffn

@deffn {} @code{bzz.deposit(amount)}
@deffnx {} {}
@end deffn


HTTP API
================

Swarm provides access through the @code{http} gateway, typically running on the local machine. This is also how queries to @code{bzz:} protocol URLs are to be handled. The following @code{http}
methods are supported:

@command{GET} provides read access to Swarm-hosted content. The first
section of the URI (until the first slash) is either @code{raw}, the
swarm hash of a manifest or a registered name which resolves to such.
Range queries are supported. URIs beginning with @code{raw/} are always
followed by a Swarm object's swarm hash. If the object exists, it is
returned in raw binary format with a generic @code{Content-type:
application/octet-stream} header. URIs beginning with the swarm hash of
a manifest object or its registered name are optionally followed by an
arbitrary URI-compatible string, which is resolved in accordance with
the content of the manifest, including the returned object's
@code{Content-type} header.

@command{PUT} is the way to insert objects into Swarm. If the URI is
@code{raw} (not followed by anything), then upon successful insertion
the swarm hash of the object is returned as @code{text/plain}. If the
URI is anything else, the returned swarm hash is that of a manifest
object in which the given URI (with the modified manifest hash, of
course) would resolve to the uploaded object, while every other URI
resolves the same way it did in the manifest object addressed by the
original URI.

@command{DELETE} is very similar to @command{PUT}, except that instead
of adding or modifying the reference in the manifest object it removes
it. Note that it does not actually delete anything from swarm (which is
conceptually impossible), merely creates a new manifest object that does
not contain the given reference.
