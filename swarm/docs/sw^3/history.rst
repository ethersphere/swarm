*******************************
Web Hosting and Incentivization
*******************************


In order to both appreciate the problems we are trying to solve and understand the demands we make on the swarm, a little history is useful.

While the Internet in general and the World Wide Web in particular
dramatically reduced the costs of disseminating information, putting
(almost) a publisher's power at (almost) every user's fingertips, these
costs are still not zero and their allocation heavily influences who
gets to publish what and who gets to enjoy what.

Web 1.0
-------------------

In the times of Web 1.0, in order to have your content accessible by the
whole world, you would typically fire up a web server or use some web
hosting (free or cheap) and upload your content that could be navigated
through a set of html pages. If your content was unpopular, you'd still
had to either maintain the server or pay the hosting to keep it
accessible, but true disaster struck when, for one reason or another, it
became popular (e.g. you got *slashdotted*). At this
point, your traffic bill skyrocketed just before either your server
crashed under the load or your hosting provider throttled your bandwidth
to the point of making your content essentially unavailable for the
majority of your audience. If you wanted to stay popular, you had to
invest in HA clusters and fat pipes and with the growth of your
popularity, your costs grew, without any obvious way to cover them.
There were very few practical ways to let (let alone *make*) your audience share the burden of information dissemination directly.

The common wisdom at the time was that it would be ISP's that would come to the rescue, since in the early days of the
Web revolution, bargaining about peering arrangements between ISP's
involved arguments about where providers and where consumers are and
which ISP is making money from the other's network. Indeed, when there
was a sharp disbalance between originators of TCP connection requests
(i.e. SYN packets), it was customary for the originator ISP to pay the
recipient ISP, which made the latter somewhat incentivized to help
hosters of popular content. In practice, however, this incentive
structure usually resulted in putting a free *pr0n* or
*warez* server in the server room to tilt the scales
of SYN packet counters. Blogs catering to a niche audience had no way of
competing and were generally left out in the cold. Note, however, that
back then, creator-publishers typically owned their content.

Web 2.0
-----------------------

The transition to Web 2.0 changed much of that. Context-sensitive
targeted advertizing offered a Faustian bargain to content producers. As
in "We give you scalable hosting that would cope with any traffic your
audience throws at it, but you give us substantial control over your
content; we are going to track each member of your audience and learn --
and own -- as much of their personal data as we can, we are going to
pick who can and who cannot see it, we are going to proactively censor
it and we may even
report on you, for the same reason. Thus, millions of small content
producers created immense value for very few corporations, getting
only peanuts (typically, free hosting) in exchange.

Peer-to-peer networks
----------------------

At the same time, however, the P2P revolution was gathering pace. Actually, P2P traffic
very soon took over the majority of packets flowing through the pipes,
quickly overtaking the above mentioned SYN-bait servers. If anything, it
proved beyond doubt that using the hitherto massively underutilized
upstream bandwidth of regular end-users, they could get the same kind of
availability and bandwidth for their content as that provided by big
corporations with data centers attached to the fattest pipes of the
internet's backbone. What's more, this could be achieved at a fraction of the cost. In particular, users retained a lot more control and freedom over their data. Finally, this mode of data distribution proved to be remarkably resilient even in the face of powerful and well-funded entities extending great efforts to shut it down.


On the other hand, even the most evolved mode of P2P file sharing, which
is trackerless Bittorrent, is just that: file sharing. It is
not suitable for providing the
kind of interactive, responsive experience that people came to expect
from web applications on Web 2.0. Simply sharing upstream bandwidth and
hard-drive space and a tiny amount of computing power without proper
accounting and indexing only gets you so far.
However, if you add to the mix a few more emergent technologies -- most importantly the blockchain -- you get what we believe to deserve the Web 3.0 moniker: a decentralized,
censorship-resistant way of sharing and even collectively creating
*interactive* content, while retaining full control over it. The price is surpisingly low and mostly consists of the resourses supplied by the super-computer (by yesteryear's standards) that you already own or can rent for peanuts.

The economics of bittorrent and its limits
----------------------------------------------------

The genious of Bittorrent lies in its clever resource optimisation:
If many clients want to download the same content from you, give them
different parts of it and let them swap the missing parts between one
another in a tit-for-tat fashion. This way, the upstream bandwidth use
of a content hoster (:dfn:`seeder` in Bittorrent parlance) is roughly the same, no matter how many clients want to download it simultaneously. This solves the most painful issue of the
ancient HTTP underpinning the World Wide Web.

Cheating (i.e. feeding your peers with garbage) is discouraged by the
use of :dfn:`hierarchical piecewise hashing`, whereby a package offered
for download is identified by a single short hash, and any part can be
cryptographically proven to be a specific part of the package without
all the other parts, with a very small overhead.

This beautifully simple approach has three main shortcomings, somewhat
related:

* There are no built-in incentives to seed downloaded content. In particular, one cannot exchange the upstream bandwidth provided by seeding one content for downstream bandwidth required for downloading some other content. Effectively, upstream bandwidth provided by seeding somebody else's content is not directly rewarded in any way.
* Typically, downloads start slowly and with delay. Clients that are further ahead in downloading have much more to offer to and much less to demand from newcomers. This results in bittorrent downloads starting as a trickle before turning into a full-blown torrent of bits. This severely limits the use of bittorrent in responsive interactive applications.
* Small chunks of data can only be shared in the context of the larger file that they are part of. We find peers sharing the content we seek by querying the Distributed Hash Table (DHT) for said file. Thus a peer sharing only part of a file needs to know what that file is in order to be found in the DHT, and conversely, if the peer doesn't know that the data chunks belong some file the peer will not be found by users seeking that file. This commonly happens for example when the same chunks of data appear verbatim in multiple files. Also, unless their objective is simply to get the missing parts of a file from their peers, nodes are not rewarded for their sharing efforts (storage and bandwidth), just like seeders.


Towards Web 3.0
-----------------------------

In order to enable responsive distributed web applications (called dapps
in Web 3.0 communities), :dfn:`IPFS` (:cite:`ipfs2014`) had to introduce a few major improvements over Bittorrent. The most immediately apparent novelty is the highly web-compatible URL-based retrieval. In addition, the directory (also organized as a DHT) has been vastly
improved, making it possible to search for any part of any file (called
*chunk*). It has also been made very flexible and pluggable in order to work with any kind of storage backend, be it a laptop with intermittent wifi, or a sophisticated HA cluster in a fiber-optic connected datacenter.

A further important innovation is that IPFS has incentivisation factored out into pluggable modules. Modules such as bitswap for example establish that it is in the interest of greedy downloaders to balance the load they impose on other nodes, and also that it is in every node's interest to host popular content. Bitswap or no bitswap, IPFS largely solves the problem of content consumers helping shouldering the costs of information dissemination.


..
  Secondly, incentivization has been factored out into pluggable modules (such as bitswap), making it possible to behave altruistically. Moreover, it is the default behavior of IPFS nodes, vastly improving performance for consumers. Because of the improved directory, it is in the interest of greedy downloaders to balance the load they impose on other nodes; unlike in the case of bittorrent, they do not need to be forced to do so. The naive default behavior of IPFS nodes is to download what they want as fast as  they can from those who provide it, while automatically caching, advertizing and uploading upon request everything they come across. They use their downstream bandwidth to the maximum extent they can, while do not limit the use of their upstream bandwidth beyond their physical limit. This, together with a few very powerful and well-connected nodes provided by the company behind IPFS, results in a very impressive performance even without any additional incentive module.

..
  One measure by which IPFS aims to shield its users from legal liability is that, just like in the case of bittorrent, there is no such thing as "pushing" anything onto an IPFS node. Sharing anything on IPFS simply means making it available on one's own node and known in the directory. However, naive consumers immediately replicate all the content they download and also make it available. Public HTTP gateways (most run by the company behind IPFS) provide automatic replication for whatever content is being accessed through them.

..
  While there is not much to gain for the user by choking uploads, or falsely advertizing content, without bitswap there is not much penalty for it either. However, bitswap incentivizes the hosting of popular content, since the constraint of swapped bits coming from the same piece of content are gone in IPFS. If you host popular content, bitswap-guarded nodes will be nice to you. There aren't that many of them, though. In this early stage of abundance, while supplied disk and bandwidth vastly outstrip demand, the system works fine as it is. If bottlenecks emerge either due to increased use or malicious intent, bitswap can be expected to become more popular as a security measure against widespread freeriding. Bitswap or no bitswap, IPFS largely solves the problem of content consumers helping shouldering the costs of information dissemination.

The same problem with lack of incentives is apparent in various other projects such as zeronet, i2p or MAIDSAFE.
Incentivization for distributed document storage is still a relatively new research field. Especially in the light of the blockchain. The Tor network has seen suggestions (:cite:`jansen2014onions`, :cite:`ghoshetal2014tor`) but these schemes are largely academic, they are not built in at the heart of the underlying system. Bitcoin has also been repurposed to drive other systems like Permacoin (:cite:`miller2014permacoin`) or Sia (:cite:`vorick2014sia`), some use their own blockchain, altcoin such as Metadisk (:cite:`wilkinson2014metadisk`) for Storj (:cite:`wilkinsonetal2014storj`) or Filecoin (:cite:`filecoin2014`) for IPFS :cite:`ipfs2014`.

What is still missing from the above incentive system, is the possibility to rent out
large amounts of disk space to those willing to pay for it, irrespective
of the popularity of their content; and conversely there is also way to deploy your interactive dynamic content to be stored in the cloud - ":dfn:`upload and disappear`".

The objective of any :index:`incentive system` for p2p content distribution is to encourage cooperative behavior and discourage freeriding: the uncompensated depletion of limited resources.
In what follows we present our current thinking for a comprehensive incentive system for swarm implemented through a suite of smart contracts. The incentive system leverages the ethereum infrastructure and the underlying value asset, :index:`Ether`.

The incentive strategy outlined here aspires to satisfy the following constraints:

* It is in the node's interest irrespective of whether other nodes follow it or not.
* It makes it expensive to hog other nodes' resources.
* It does not impose unreasonable overhead.
* It plays nice with "naive" nodes.
* It rewards those that play nice, including those following this strategy.

In the context of swarm, storage and bandwidth are the two most important limited resources and this is reflected in our incentive scheme. The incentives for bandwith use are designed to achieve speedy and reliable data provision while the storage incentives are designed to ensure long term data preservation, ideally solving the "upload and disappear" problem. In the following section we introduce the basic functioning of the swarm incentive system and the Swarm Accounting Protocol (SWAP) which handles compensation for bandwidth use in realtime. In  :ref:`Section 3 ƒ<sec:storage>` we turn to the problem of data preservation and offer a solution.


