pragma solidity ^0.4.0;
contract SimplestENS {
    event ContentChanged(bytes32 indexed node, bytes32 hash);
    mapping(bytes32=>bytes32) records;
    function content(bytes32 node) public constant returns (bytes32 ret) {
        ret = records[node];
    }
    function setContent(bytes32 node, bytes32 hash) public {
        records[node] = hash;
        ContentChanged(node, hash);
    }
}
