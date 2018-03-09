pragma solidity 0.4.18;

import 'RLP.sol';

// SafeMath.sol
function mul(uint256 a, uint256 b)
        internal
        pure
        returns (uint256)
    {
        if (a == 0) {
        return 0;
        }
        uint256 c = a * b;
        assert(c / a == b);
        return c;
    }

    function div(uint256 a, uint256 b)
        internal
        pure
        returns (uint256)
    {
        // assert(b > 0); // Solidity automatically throws when dividing by 0
        uint256 c = a / b;
        // assert(a == b * c + a % b); // There is no case in which this doesn't hold
        return c;
    }

    function sub(uint256 a, uint256 b)
        internal
        pure
        returns (uint256) 
    {
        assert(b <= a);
        return a - b;
    }

    function add(uint256 a, uint256 b)
        internal
        pure
        returns (uint256) 
    {
        uint256 c = a + b;
        assert(c >= a);
        return c;
    }


/**
 * @title Bytes operations
 *
 * @dev Based on https://github.com/GNSPS/solidity-bytes-utils/blob/master/contracts/BytesLib.sol
 */
function slice(bytes _bytes, uint _start, uint _length)
        internal
        pure
        returns (bytes)
{
        
        bytes memory tempBytes;
        
        assembly {
            tempBytes := mload(0x40)
            
            let lengthmod := and(_length, 31)
            
            let mc := add(tempBytes, lengthmod)
            let end := add(mc, _length)
            
            for {
                let cc := add(add(_bytes, lengthmod), _start)
            } lt(mc, end) {
                mc := add(mc, 0x20)
                cc := add(cc, 0x20)
            } {
                mstore(mc, mload(cc))
            }
            
            mstore(tempBytes, _length)
            
            //update free-memory pointer
            //allocating the array padded to 32 bytes like the compiler does now
            mstore(0x40, and(add(mc, 31), not(31)))
        }
        
        return tempBytes;
    }
/**
   * @dev Recover signer address from a message by using his signature
   * @param hash bytes32 message, the hash is the signed message. What is recovered is the signer address.
   * @param sig bytes signature, the signature is generated using web3.eth.sign()
   */

function recover(bytes32 hash, bytes sig)
        internal
        pure
        returns (address)
{
        bytes32 r;
        bytes32 s;
        uint8 v;

        //Check the signature length
        if (sig.length != 65) {
        return (address(0));
        }

        // Divide the signature in v, r, and s variables
        assembly {
        r := mload(add(sig, 32))
        s := mload(add(sig, 64))
        v := byte(0, mload(add(sig, 96)))
        }

        // Version of signature should be 27 or 28, but 0 and 1 are also possible versions
        if (v < 27) {
        v += 27;
        }

        // If the version is correct return the signer address
        if (v != 27 && v != 28) {
        return (address(0));
        } else {
        return ecrecover(hash, v, r, s);
        }
}

// 4 sigs in "sigs" byte array (260 bytes)
//  sig1 [0:65]  -- signed transaction txhash, must match confSig
//  sig2 [65:130] -- signed transaction txhash, must match confSig2
//  confSig1 [130:195] -- for confirmationHash
//  confSig2 [195:196+65]
// where confirmationHash = hash(txhas, sig1, sig2, roothash)

// inputCount is based on blocknum1 and blocknum2
function checkSigs(bytes32 txHash, bytes32 rootHash, uint256 inputCount, bytes sigs)
        internal
        view
        returns (bool)
{
        require(sigs.length % 65 == 0 && sigs.length <= 260);
        bytes memory sig1 = ByteUtils.slice(sigs, 0, 65);
        bytes memory sig2 = ByteUtils.slice(sigs, 65, 65);
        bytes memory confSig1 = ByteUtils.slice(sigs, 130, 65);
        bytes32 confirmationHash = keccak256(txHash, sig1, sig2, rootHash);
        if (inputCount == 0) {
            return msg.sender == ECRecovery.recover(confirmationHash, confSig1);
        }
        if (inputCount < 1000000) { // only blocknum1
            return ECRecovery.recover(txHash, sig1) == ECRecovery.recover(confirmationHash, confSig1);
        } else {
            bytes memory confSig2 = ByteUtils.slice(sigs, 195, 65);
            bool check1 = ECRecovery.recover(txHash, sig1) == ECRecovery.recover(confirmationHash, confSig1);
            bool check2 = ECRecovery.recover(txHash, sig2) == ECRecovery.recover(confirmationHash, confSig2);
            return check1 && check2;
        }
}

function max(uint256 a, uint256 b)
        internal
        pure
        returns (uint256)
{
        if (a > b) 
            return a;
        return b;
}

function checkMembership(bytes32 leaf, uint256 index, bytes32 rootHash, bytes proof)
        internal
        pure
        returns (bool)
{
        require(proof.length == 512);
        bytes32 proofElement;
        bytes32 computedHash = leaf;

        for (uint256 i = 32; i <= 512; i += 32) {
            assembly {
                proofElement := mload(add(proof, i))
            }
            if (index % 2 == 0) {
                computedHash = keccak256(computedHash, proofElement);
            } else {
                computedHash = keccak256(proofElement, computedHash);
            }
            index = index / 2;
        }
        return computedHash == rootHash;
}

contract PriorityQueue {
    using SafeMath for uint256;

    /*
     *  Modifiers
     */
    modifier onlyOwner() {
        require(msg.sender == owner);
        _;
    }

    /* 
     *  Storage
     */
    address owner;
    uint256[] heapList;
    uint256 public currentSize;

    function PriorityQueue()
        public
    {
        owner = msg.sender;
        heapList = [0];
        currentSize = 0;
    }

    function insert(uint256 k) 
        public
        onlyOwner
    {
        heapList.push(k);
        currentSize = currentSize.add(1);
        percUp(currentSize);
    }

    function minChild(uint256 i)
        public
        view
        returns (uint256)
    {
        if (i.mul(2).add(1) > currentSize) {
            return i.mul(2);
        } else {
            if (heapList[i.mul(2)] < heapList[i.mul(2).add(1)]) {
                return i.mul(2);
            } else {
                return i.mul(2).add(1);
            }
        }
    }

    function getMin()
        public
        view
        returns (uint256)
    {
        return heapList[1];
    }

    function delMin()
        public
        onlyOwner
        returns (uint256)
    {
        uint256 retVal = heapList[1];
        heapList[1] = heapList[currentSize];
        delete heapList[currentSize];
        currentSize = currentSize.sub(1);
        percDown(1);
        return retVal;
    }

    function percUp(uint256 i) 
        private
    {
        while (i.div(2) > 0) {
            if (heapList[i] < heapList[i.div(2)]) {
                uint256 tmp = heapList[i.div(2)];
                heapList[i.div(2)] = heapList[i];
                heapList[i] = tmp;
            }
            i = i.div(2);
        }
    }

    function percDown(uint256 i)
        private
    {
        while (i.mul(2) <= currentSize) {
            uint256 mc = minChild(i);
            if (heapList[i] > heapList[mc]) {
                uint256 tmp = heapList[i];
                heapList[i] = heapList[mc];
                heapList[mc] = tmp;
            }
            i = mc;
        }
    }
}

contract RootChain {
    using SafeMath for uint256;
    using RLP for bytes;
    using RLP for RLP.RLPItem;
    using RLP for RLP.Iterator;
    using Merkle for bytes32;

    /*
     * Events
     */
    event Deposit(address depositor, uint256 amount);

    /*
     *  Storage
     */
    /*
    childChain: A list of Plasma blocks, for each block storing (i) the Merkle root, (ii) the time the Merkle root was submitted.
    */
    mapping(uint256 => childBlock) public childChain;

    /*
    A list of submitted exit transactions, storing
     (i) the submitter address, and 
     (ii) the UTXO position (Plasma block number, txindex, outindex). 
    This must be stored in a data structure that allows transactions to be popped from the set in order of priority.
   */
    mapping(uint256 => exit) public exits;
    mapping(uint256 => uint256) public exitIds;
    PriorityQueue exitsQueue;

    // owner (set at initialization time)
    address public authority;

    
    uint256 public currentChildBlock;
    uint256 public lastParentBlock;
    uint256 public recentBlock;   // not used
    uint256 public weekOldBlock;  // updated in incrementOldBlocks [when submitBlock]

    struct exit {
        address owner;
        uint256 amount;
        uint256[3] utxoPos;
    }

    struct childBlock {
        bytes32 root;
        uint256 created_at;
    }

    /*
     *  Modifiers
     */
    modifier isAuthority() {
        require(msg.sender == authority);
        _;
    }

    modifier incrementOldBlocks() {
        while (childChain[weekOldBlock].created_at < block.timestamp.sub(1 weeks)) {
            if (childChain[weekOldBlock].created_at == 0) 
                break;
            weekOldBlock = weekOldBlock.add(1);
        }
        _;
    }

    function RootChain()
        public
    {
        authority = msg.sender;
        currentChildBlock = 1;
        lastParentBlock = block.number;
        exitsQueue = new PriorityQueue();
    }

    /* 

    Plasma Chain block submission: A Plasma block can be created in
    one of two ways. First, the operator of the Plasma chain can
    create blocks. Second, anyone can deposit any quantity of ETH into
    the chain, and when they do so the contract adds to the chain a
    block that contains exactly one transaction, creating a new UTXO
    with denomination equal to the amount that they deposit.

    Each Merkle root should be a root of a tree with depth-16 leaves, where each leaf is a transaction. 

    A transaction is an RLP-encoded object of the form:

      [blknum1, txindex1, oindex1, sig1, # Input 1
       blknum2, txindex2, oindex2, sig2, # Input 2
       newowner1, denom1,                # Output 1
       newowner2, denom2,                # Output 2
       fee]

    Each transaction has 2 inputs and 2 outputs, and the sum of the denominations of the outputs plus the fee must equal the sum of the denominations of the inputs. 
    
    The signatures must be signatures of all the other fields in the transaction, with the private key corresponding to the owner of that particular output. 

    A deposit block has all input fields, and the fields for the second output, zeroed out. 
    To make a transaction that spends only one UTXO, a user can zero out all fields for the second input.
    */
    function submitBlock(bytes32 root)
        public
        isAuthority
        incrementOldBlocks
    {
        require(block.number >= lastParentBlock.add(6));
        childChain[currentChildBlock] = childBlock({
            root: root,
            created_at: block.timestamp
        });
        currentChildBlock = currentChildBlock.add(1);
        lastParentBlock = block.number;
    }

    /*
     generates a block that contains only one transaction, generating a new UTXO into existence with denomination equal to the msg.value deposited
      txList[0-5, 9]: 0
      txList[6]: toAddress
      txList[7]: msg.value
     */
    function deposit(bytes txBytes)
        public
        payable
    {
        var txList = txBytes.toRLPItem().toList();
        require(txList.length == 11);
        for (uint256 i; i < 6; i++) {
            require(txList[i].toUint() == 0);
        }
        require(txList[7].toUint() == msg.value); // has to match!
        require(txList[9].toUint() == 0);
        bytes32 zeroBytes;
        // generate root through a lot of Keccak hashing
        bytes32 root = keccak256(keccak256(txBytes), new bytes(130));
        for (i = 0; i < 16; i++) {
            root = keccak256(root, zeroBytes);
            zeroBytes = keccak256(zeroBytes, zeroBytes);
        }
        // make a new block 
        childChain[currentChildBlock] = childBlock({
            root: root,
            created_at: block.timestamp
        });
        currentChildBlock = currentChildBlock.add(1);
        Deposit(txList[6].toAddress(), txList[7].toUint());
    }

    function getChildChain(uint256 blockNumber)
        public
        view
        returns (bytes32, uint256)
    {
        return (childChain[blockNumber].root, childChain[blockNumber].created_at);
    }

    function getExit(uint256 priority)
        public
        view
        returns (address, uint256, uint256[3])
    {
        return (exits[priority].owner, exits[priority].amount, exits[priority].utxoPos);
    }

    /*
    startExit(uint256 plasmaBlockNum, uint256 txindex, uint256 oindex, bytes tx, bytes proof, bytes confirmSig): 
     starts an exit procedure for a given UTXO. Requires as input
     (i)  the Plasma block number (txPos[0]) and tx index in which the UTXO was created, 
     (ii) the output index (txPos[2]), 
     (iii) the transaction containing that UTXO (txPos[1]), 
     (iv) a Merkle proof of the transaction (proof)
     (v) a confirm signature (sig) from each of the previous owners of the now-spent outputs that were used to create the UTXO.
    */
    function startExit(uint256[3] txPos, bytes txBytes, bytes proof, bytes sigs)
        public
        incrementOldBlocks
    {
        var txList = txBytes.toRLPItem().toList();
/*
    A transaction is an RLP-encoded object (length 11) of the form: (sig1, sig2 is put in "sigs")

      [0: blknum1, 1: txindex1, 2: oindex1, (sig1) # Input 1 
       3: blknum2, 4: txindex2, 4: oindex2, (sig2) # Input 2
       6: newowner1, denom1,                 # Output 1
       8: newowner2, denom2,                 # Output 2
       10: fee]
*/
        require(txList.length == 11);
        require(msg.sender == txList[6 + 2 * txPos[2]].toAddress());  // sender has to be newowner1 (0) or newowner2 (1) depending on what txPos[2]  is
        bytes32 txHash = keccak256(txBytes);
        bytes32 merkleHash = keccak256(txHash, ByteUtils.slice(sigs, 0, 130));  // txHash, sig1, sig2
        uint256 inputCount = txList[3].toUint() * 1000000 + txList[0].toUint();
        require(Validate.checkSigs(txHash, childChain[txPos[0]].root, inputCount, sigs));

        txPos: 
        // function checkMembership(bytes32 leaf, uint256 index, bytes32 rootHash, bytes proof)
        // leaf: txPos[1]
        // index: childChain[txPos[0]].root
        // proof: supplied 512-byte (32*16) 
        require(merkleHash.checkMembership(txPos[1], childChain[txPos[0]].root, proof));
        // arrange exits into a priority queue structure, where priority is normally the tuple (blknum, txindex, oindex) (alternatively, blknum * 1000000000 + txindex * 10000 + oindex). 
        // txPos[0] - blocknum
        // txPos[1] - txindex
        // txPos[2] - oindex
        uint256 priority = 1000000000 + txPos[1] * 10000 + txPos[2];
        uint256 exitId = txPos[0].mul(priority);
        priority = priority.mul(Math.max(txPos[0], weekOldBlock));
        require(exitIds[exitId] == 0);
        require(exits[priority].amount == 0);
        exitIds[exitId] = priority;
        exitsQueue.insert(priority);
        /*
        However, if when calling exit, the block that the UTXO was created in is more than 7 days old, then the blknum of the oldest Plasma block that is less than 7 days old is used instead. There is a
        passive loop that finalizes exits that are more than 14 days old, always processing exits in order of priority (earlier to later).

        This mechanism ensures that ordinarily, exits from earlier UTXOs are processed before exits from older UTXOs, 
        and particularly, if an attacker makes a invalid block containing bad UTXOs, the holders of all earlier UTXOs 
        will be able to exit before the attacker. 

        The 7 day minimum ensures that even for very old UTXOs, there is ample time to challenge them.
        */
        exits[priority] = exit({
            owner: txList[6 + 2 * txPos[2]].toAddress(),
            amount: txList[7 + 2 * txPos[2]].toUint(),
            utxoPos: txPos
        });
    }

    /*
    challengeExit(uint256 exitId, uint256 plasmaBlockNum, uint256 txindex, uint256 oindex, bytes tx, bytes proof, bytes confirmSig): 
    challenges an exit attempt in process, by providing a proof that the TXO was spent, 
    the spend was included in a block, and the owner made a confirm signature.
    */
    function challengeExit(uint256 exitId, uint256[3] txPos, bytes txBytes, bytes proof, bytes sigs, bytes confirmationSig)
        public
    {
        var txList = txBytes.toRLPItem().toList();
        require(txList.length == 11);
        uint256 priority = exitIds[exitId];
        uint256[3] memory exitsUtxoPos = exits[priority].utxoPos;
        require(exitsUtxoPos[0] == txList[0 + 2 * exitsUtxoPos[2]].toUint());
        require(exitsUtxoPos[1] == txList[1 + 2 * exitsUtxoPos[2]].toUint());
        require(exitsUtxoPos[2] == txList[2 + 2 * exitsUtxoPos[2]].toUint());
        var txHash = keccak256(txBytes);
        var confirmationHash = keccak256(txHash, sigs, childChain[txPos[0]].root);
        var merkleHash = keccak256(txHash, sigs);
        address owner = exits[priority].owner;
        require(owner == ECRecovery.recover(confirmationHash, confirmationSig));
        require(merkleHash.checkMembership(txPos[1], childChain[txPos[0]].root, proof));

        delete exits[priority];
        delete exitIds[exitId];
    }

    function finalizeExits()
        public
        incrementOldBlocks
        returns (uint256)
    {
        uint256 twoWeekOldTimestamp = block.timestamp.sub(2 weeks);
        exit memory currentExit = exits[exitsQueue.getMin()];
        while (childChain[currentExit.utxoPos[0]].created_at < twoWeekOldTimestamp && exitsQueue.currentSize() > 0) {
            // return childChain[currentExit.utxoPos[0]].created_at;
            uint256 exitId = currentExit.utxoPos[0] * 1000000000 + currentExit.utxoPos[1] * 10000 + currentExit.utxoPos[2];
            currentExit.owner.transfer(currentExit.amount);
            uint256 priority = exitsQueue.delMin();
            delete exits[priority];
            delete exitIds[exitId];
            currentExit = exits[exitsQueue.getMin()];
        }
    }
}
