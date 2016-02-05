*******************
Introduction
*******************

=================
Swarm
=================

..  * extention allows for per-format preference for image format

..  image:: img/swarm-logo.jpg
   :height: 300px
   :width: 300 px
   :scale: 50 %
   :alt: swarm-logo
   :align: right


This document presents @value{SWARM} which is being developed as part
of the crypto 2.0 vision for web 3.0.

Swarm is a decentralised document storage solution. It is deeply integrated within
ethereum and provides one of the most prominent base layer services: data storage
and content distribution. It is free and self-sustaining via its incentive structure
implemented as ethereum smart contracts.

This document provides you with:


Background
=================

The primary objective of Swarm is to provide a sufficiently
decentralized and redundant store of Ethereum's public record, in
particular to store and distribute Đapp code and data as well as
block chain data.

From an economic point of view, it allows participants to efficiently
pool their storage and bandwidth resources in order to provide the
aforementioned services to all participants.

These objectives entail the following design requirements:

* distributed storage, inclusivity, long tail of power law
* flexible expansion of space without hardware investment decisions, unlimited growth
* zero downtime
* immutable unforgeable verifiable yet plausibly deniable
* no single point of failure, fault and attack resilience
* censorship resistance, permanent public record
* self-managed sustainability via incentive system
* efficient market driven pricing. tradeable trade off of memory, persistent storage, bandwidth (and later computation)
* efficient use of the blockchain by swarm accounting protocol
* deposit-challenge based guaranteed storage

About
===================

This document
---------------------

This document source code is found at @url{https://github.com/ethersphere/swarm/tree/master/book}
The most uptodate swarm book in various formats is available on the old web
@url{http://ethersphere.org/swarm/docs} as well as on swarm @url{bzz://swarm/docs}


Status
---------------

The status of swarm is proof of concept vanilla prototype tested on toy network.
It is highly experimental code and untested in the wild.

License
-------------

Swarm is free software.

It is licensed under @dfn{LGPL}, which roughly means the following.

There are @emph{no restrictions on downloading} it other than
your bandwidth and our slothful ways of making things available.

There are @emph{no restrictions on use} either other than its deficiencies,
clumsy features and outragous bugs. However, this can be amended,
because there are @i{no restrictions on modifying} it either.
See also @ref{Contributing}.

Freedom of use implies that anything goes.

What is more, there are @i{no restrictions on redistributing} this software or
any modified version of it.

For some legalese telling you the same, read the License @c
@uref{http://creativecommons.org/licenses/LGPL/2.1/}

@c Creative Commons.

@c @ref{Creative Commons}.

Credits
---------------------

Swarm is code by Ethersphere (ΞTHΞЯSPHΞЯΞ), an alliance of ethdev developers: Viktor Trón, Dániel A. Nagy and Zsolt Felföldi.

Swarm is funded by the Ethereum Foundation.

Special thanks to

* Felix Lange, Alex Leverington for inventing and implementing devp2p/rlpx;
* Jeffrey Wilcke and the go team for continued support, testing and direction;
* Gavin Wood and Vitalik Buterin for the vision;
* Alex Van der Sande, Fabian Vogelsteller and Dániel Varga for a lot of inspiring discussions and ideas, shaping design from early on;
* Aron Fischer for his ideas and hands-on help with analysis, documentation and testing;
* Roman Mandeleil and Anton Nashatyrev for the java implementation;

Community
-------------------

* Gitter: https://gitter.im/ethereum/swarm
* Reddit: http://www.reddit.com/r/ethereum
* Forum: https://forum.ethereum.org/categories/swarm

Reporting a bug
-------------------

Issues are tracked on github and github only: @url{https://github.com/ethereum/go-ethereum/labels/swarm}

See the ethereum developer's guide for how to submit a bug report, feature request or fix: https://github.com/ethereum/go-ethereum/wiki/Developers'-Guide

Contributing
--------------------

See the ethereum developer's guide for how to contribute to the project. https://github.com/ethereum/go-ethereum/wiki/Developers'-Guide

Roadmap
-------------------

For actual issues, see https://github.com/ethereum/go-ethereum/labels/swarm
* SWAP^3: swarm accounting protocol stage 3 adding debt swap (accreditation)
* SWEAR & SWINDLE storage incentives: receipts and litigation
* SWORD ethereum blockchain, state, contract storage, logs and receipts on swarm
* network stress testing, viability, scalability
* latency and traffic simulations for routing
* encryption for basic PD masking
* proveable prefix array for full text search,
* swarm db, swarm fs via fuse

Resources
----------------

Talks:

* Dr. Daniel A. Nagy: Keeping the Public Record Safe and Accessible. Ethereum ÐΞVCON0, Berlin. 2014 - @url{https://www.youtube.com/watch?v=QzYZQ03ON2o}
* Viktor Trón, Daniel A. Nagy: Swarm. ÐΞVCON1, London. 2015

