*****************************************
Swap - Swarm Accounting Protocol
*****************************************

SWAP is the abbreviation for the *Swarm Accounting Protocol*. It is the protocol by which peers in the Swarm network keep track of chunks delivered and received and the resulting (micro-)payments owed. On its own, SWAP can function in a wider context too and is presented here as a generic micropayment scheme suited for pairwise accounting between peers. And yet, while generic by design, our first use of it is for accounting of bandwidth and storage as part of the incentivisation of data transfer and ensured archival in the Swarm decentralised peer to peer storage network.

..  index::
   Swarm Accounting Protocol (SWAP)
   incentivisation
   micropayments


There are three levels of SWAP:

* ``Swap^1``: Swarm Accounting Protocol
* ``Swap^2``: Swarm Accounting Protocol with Swift Automatic  Payment
* ``Swap^3``: Swarm Accounting Protocol with Swift Automatic  Payments  and Debt Swap

SWAP
==================

In the context of Swarm, SWAP will be used as an accounting protocol for exchanged chunks of data. However in general SWAP is a protocol (in the broad sense as 'scheme of interaction') of accounting that enables direct trading of any commodity class that

* is immediately quantifiable
* is typically mutually provided and used, but with arbitrary variance
* is used over a long period
* has a value which is strictly additive (quantity vs value is linear)
* has source-independent value (exchange of equal quantities is always fair)
* has relatively stable cost

A SWAP instance is meant to be linked to each pair of peers. Typically a peer would be running a client on some network, where the commodity is exchanged. SWAP also allows peers to be widely variable in their ability or desire to use or provide service. In such cases the provider gets compensation for the service, i.e., peers can trade service for a token (means of exchange).

In the context of the Swarm network, the node at either end of a peer connection would be running its own instance of SWAP, accounting for chunks exchanged with the peer on the other side of the connection.

Peers in the Swarm network will typically exchange chunks of data. Accounting one chunk of data passed one way balanced by chunks of data passing the other way is costless and therefore preferable to paying for each chunk via a (micro-)transaction on the ethereum blockchain, either individually or in aggregate. Stated more generally we can say that accounting service for service is costless since the transfer of the service takes place between the peers. Compensation with a token however has a cost (at the very least it involves an extra transaction to be processed by trusted third party or a decentralised consensus network).

..  index::
  chunk

Note that in the context of a micropayment solution, it is implicit that immediate payment on each transaction is not viable. If the service events to account for are cheap and numerous, transaction cost for compensation will dominate making direct verified payments both prohibitively slow and expensive.

Thus we have established that we cannot feasably pay for each chunk as it is transferred and if chunks are transferred more in one direction along a peer connection than in the other, then one side must therefore be allowed to accrue debts with respect to the other. This in turn implies that cheating (not paying debts) is a real possibility and we must devise schemes to combat it.

..  index::
   cheating

For example, such schemes could include up-front deposits in smart contracts and punishment for non-compliance. Each participant would have to lock substantial funds on the blockchain in order to participate in the swarm network. This protects nodes from insolvency of their peers, but it is not ideal and affects liquidity of capital. Instead we rely on a reputation system explained below.


Any solution of the micropayment problem which uses *delayed payment* instruments, immediately leads to the question of debtors and how much debt we can afford; or conversely, how much upfront deposit we require to establish trust. For now we make the simplifying assumption that debt collection by external agents cannot be relied on and therefore the only action available is to minimize the risk and cut business in case of fraud or insolvency. In the context of a peer to peer network of nodes where SWAP is on active connections, this translates to *disconnecting* (or *dropping*) the peer. The penalising effect of dropping a peer can be increased by permanent banning and blacklisting or any measure that persists the trustworthiness of peers beyond one uptime session.

..  index::
  insolvency
  disconnection
  delayed payment
  debt

i.e. if a node misbehaves, it is dropped from the network.

In this context then we must decide on reasonable limits of debts accrued. What is our tolerance for debt?

:dfn:`Tolerance` is defined as the threshold that triggers dropping a peer and it is given as a limit on service debt.

..  index:: tolerance

If tolerance is set too low, the resulting (unintentional?) disconnects can hurt the network especially if reconnecting to new peers is costly, or not possible because new peers are hard to come by.

If there are many peers available to provide and consume and it is free to switch, then peers can afford low tolerance.
If peers are scarce or it is costly to switch, tolerance has to be high.
This introduces potential risk if tolerance is not checked against locked funds. If the amount is not secured (i.e., if solvency is not guaranteed), the network needs to rely on reputation as an incentive.

..  index:: reputation

In Swarm, we develop a system of deferred payment by promisory notes -- somewhat analogous to paying by cheque. The cheque payments are numerous and immediate, but the cashing-in of cheques is expected to be rare resulting in fewer blockchain transactions needed overall.

..  index:: cheque

In general, SWAP allows any type of *payment system* that has an *issue* and
*receive* method. While issue and receive could implement immediate payment (i.e, sending, receiving/confirming a transaction) in this system we stipulate that issue results in a *3rd party proveable promise of payment* which can be cashed (to the beneficiary) using the payment processor.

..  index:: cheque book

In the Swarm cheque book, each new cheque incorporates and extends the previous cheques issued, so that only one cheque (the last one) ever needs to be cashed in. A peer can choose to cash in often (if the counterparty is not trusted), or cash in rarely (to save on transaction costs and as a side effect, to minimise number of transactions needed overall).

Thus an immediate cut in transaction costs is possible using promises that record a cumulative tally of debt (i.e. my last promise includes earlier promises). The payment processors record the cumulative amount in the event of cashing out on a promise. Next time a promise is shown, only the difference is actually paid. This makes it possible for the beneficiary to skip cashing out on some payments or even cash out only when necessary (e.g., in need of funds or avert high risk).

Delaying cashing out by the beneficiary is a crucial feature.  Since transactions cost money, there is an incentive to minimize their number. Delaying cashing out does exactly that.
But letting uncashed payments grow comes with growing risk, namely bigger loss in the event of insolvency.

..  index:: insolvency

To reduce insolvency risks we could rely on a *reputation system*. In the Swarm network, business is only conducted over the ethereum devp2p connections; crucially, these connections are long term, thus allowing a trust network to develop. For example, we can cash all incoming cheques immediately when we are first connected to a peer, but relax once we have established that a peer is behaving well.

..  index:: reputation system

Thus saving on transaction costs while managing risk is possible if participants are incentivised to conduct continual business as the same identity (basically a reputation system). This is possible because long term participation encourages compliance (*discipline of repeated transactions*), since honest users set the standard which can then be expected. Naive new nodes need to provide the service first before using it or else need to pay their way in for not working.

SWAP^2
=============

SWAP^2 stands for Swarm Accounting Protocol with Swift Automatic Payments.

..  index::
  Swarm Accounting Protocol (SWAP)
  autopayment
  micropayment

Our strategy as a participating node might be that newly connected peers must pay often, but older (and thus more trusted) nodes can accrue higher debts before settlement becomes necessary.

SWAP^2 allows for an enhanced automated version of SWAP which reduces transaction cost without overhead or adverse impact on security.
It is an extenstion which provides an API for setting and resetting trigger thresholds.

As a seller you can set a limit on maximum service debt and as a buyer you can set a threshold that triggers a payment.  You can also cash out automatically triggered by a limit on maximum uncashed revenue from a peer with a fallback to time period (interval after which promise is always cashed).

..  index::
   PayAt, payment threshold)
   payment threshold (``PayAt``)
   DropAt, disconnect threshold)
   disconnect threshold (```DropAt``)
   SellAt, offered price)
   offered price (``SellAt``)
   BuyAt, highest accepted price)
   highest accepted price (``BuyAt``)


SWAP^2 allows  dynamic resetting of trigger thresholds and intervals on a per-peer basis, which makes it possible to implement sophisticated autopayment strategies that process information about locked funds, reputation, credit history, insolvency etc based on which the tolerance levels can be set dynamically.

In particular, a strategy that tracks reputation (or a combination of reputation and amount of locked funds) and adjusts delay accordingly is sound in as much as risk assessment is based on creditworthiness. On the one hand, if there is no trust each promise is immediately cashed. Conversely, with unlimited trust we can infinitely postpone cashing until we actually need funds.


As a buyer you can set the limit at which you deposit funds and what is the maximum amount you keep as a *credit buffer* (which mitigates the risk of insolvency). Alternatively or in conjuction as a fallback you can set a time interval after which funds are sent to the sending contract to restore the desired credit buffer. In this scenario the sending contract is considered a type of hot wallet.
Honest users consider  the balance on the sending contract locked.
Setting autodeposit strategy manages the tradeoff between spending on transactions or invest in earning trust by a higher buffer which effectively models time preference in service use. If a peer has no reputation or deposit locked, seller will typically autocash on each IOU received. If the peer is short on funds, autodeposit on each payment is the only option.  Such a system incentivises honest use since reputation can save the transaction cost.

..  index::
   credit buffer (``AutoDepositBuffer``)
   autocash threshold (``AutoCashThreshold``)
   autodeposit threshold (``AutoDepositThreshold``)
   autocash interval (``AutoCashInterval``)
   autodeposit interval (``AutoDepositInterval``)
   AutoDepositBuffer, credit buffer
   AutoCashThreshold, autocash threshold
   AutoDepositThreshold: autodeposit threshold
   AutoCashInterval, autocash interval
   AutoCashBuffer, autocash target credit buffer)

This system is completely flexible even allowing capping of service related spending while allowing unlimited consumption via service-for-service exchange.

SWAP^3
=========

SWAP^3 further reduces transaction costs by introducing payment by IOU (*debt swap*)

..  index:: debt swap

What this means is the following. Suppose node A owes node B for N chunks and node A sends a cheque to node B over N chunks; further suppose that in the following node B receives k chunks from node A and thus owes A payment for k chunks. What a debt swap arrangement allows is for node B to certify to node A that the cheque it holds shall now be of a total value covering just N-k chunks.
