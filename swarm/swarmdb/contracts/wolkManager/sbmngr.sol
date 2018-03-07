pragma solidity ^0.4.19;

//Obsolete On-Chain Manager

library SafeMath {
    function mul(uint256 a, uint256 b) internal pure returns (uint256) {
        if (a == 0) {
          return 0;
        }
        uint256 c = a * b;
        assert(c / a == b);
        return c;
    }
    
    function div(uint256 a, uint256 b) internal pure returns (uint256) {
        uint256 c = a / b;
        return c;
    }
    
    function sub(uint256 a, uint256 b) internal pure returns (uint256) {
        assert(b <= a);
        return a - b;
    }
    
    function add(uint256 a, uint256 b) internal pure returns (uint256) {
        uint256 c = a + b;
        assert(c >= a);
        return c;
    }
} 

contract Owned {
    address public owner;
    address public newOwner;
    modifier onlyOwner { require(msg.sender == owner); _; }
    event OwnerUpdate(address _prevOwner, address _newOwner);

    function Owned() public {
        owner = msg.sender;
    }

    function transferOwnership(address _newOwner) public onlyOwner {
        require(_newOwner != owner);
        newOwner = _newOwner;
    }

    function acceptOwnership() public {
        require(msg.sender == newOwner);
        OwnerUpdate(owner, newOwner);
        owner = newOwner;
        newOwner = 0x0;
    }
}

// ERC20 Interface
contract ERC20 {
    function totalSupply() public constant returns (uint _totalSupply);
    function balanceOf(address _owner) public constant returns (uint balance);
    function transfer(address _to, uint _value) public returns (bool success);
    function transferFrom(address _from, address _to, uint _value) public returns (bool success);
    function approve(address _spender, uint _value) public returns (bool success);
    function allowance(address _owner, address _spender) public constant returns (uint remaining);
    event Transfer(address indexed _from, address indexed _to, uint _value);
    event Approval(address indexed _owner, address indexed _spender, uint _value);
}

contract SBMNGR is Owned {

    struct Chunk {
        address buyer;
        uint8 claimed;
        uint8 index;
        address[] farmerList;
        uint256 declaredBlockN;
        uint256 reward;
    }
    
    struct ChunkList {
        bytes32 blockH;
        bytes32[] chunks;
    }

    
    using SafeMath for uint256;
    mapping (bytes32 => Chunk)  public claims;
    mapping (uint256 => ChunkList) chunkList;
    mapping (address => uint256) public unpaidCost;
    mapping (address => uint256) public unpaidEarning;
    mapping (address => bool) public validators;
    address public tokenAddress;
    address public tokenReserve;

    modifier onlyValidators { require(msg.sender == owner || validators[msg.sender]); _; }

    uint public rewardDifficulty = 16;
    uint public maxClaimPerChunk = 8;
    //uint public chequeAllowance = 10 ** 18;
    

    event Grant(address indexed _ValidatorAdded);
    event Revoke(address indexed _validatorRemoved, address indexed _request);    
    event Claim(address indexed _chunkOwner, address indexed _farmer, bytes32 _claim, uint _submittedBlock);
    event Reward(address indexed _chunkOwner, address indexed _farmer, address indexed _validator, bytes32 _claim, uint _settledBlock);
    event Rejection(address indexed _chunkOwner, address indexed _farmer, address indexed _validator, bytes32 _claim, uint _settledBlock);

    function setResolver(address _tokenAddr, address _reseveAddr) onlyOwner public {
        tokenAddress = _tokenAddr;
        tokenReserve = _reseveAddr;
    }
    
    function setDifficulty(uint256 _rewardDifficulty) onlyOwner public {
        // max of 256
        require(_rewardDifficulty <= 256);
        rewardDifficulty = _rewardDifficulty;
    }
    
    function addValidator(address _newValidator) public onlyOwner {
        validators[_newValidator] = true;
        Grant(_newValidator);
    }
    
    function removeValidator(address _validator) public onlyValidators {
        if(msg.sender != owner) _validator = msg.sender;
        validators[_validator] = false;
        Revoke(_validator, msg.sender);
    }
    
    function getChunks(uint _blockNumber) public view returns (bytes32[] _chunks) {
        return chunkList[_blockNumber].chunks;
    }

    function claimCount(bytes32 _chunk) public view returns (uint8 _claimedCount) {
        return claims[_chunk].claimed;
    }
    
    function getBlockHash(uint256 _blockNumber) public view returns (bytes32 _blockHash) {
        return block.blockhash(_blockNumber);
    }
    

    function getFarmers(bytes32 _chunk) public view returns(address[] farmers) {
        return claims[_chunk].farmerList;
    }
    
    function getFarmerAtIndex(bytes32 _chunk, uint8 index) public view returns(address) {
        if  (claims[_chunk].farmerList.length > 0) {
            return claims[_chunk].farmerList[index];
        }else{
            return 0x0;
        }
        
    }

    /*
    function removeFarmer(bytes32 _chunk, uint index) public returns(address[] farmers) {
        delete claims[_chunk].farmerList[index];
        return claims[_chunk].farmerList;
    }
    */

    function addFarmer(bytes32 _chunk, address _farmerToAdd) internal returns (bool) {
        var farmers = claims[_chunk].farmerList;
        for (uint i = 0; i< farmers.length; i++){
            if(farmers[i] == _farmerToAdd){
                return false;
            }
        }
        claims[_chunk].farmerList.push(_farmerToAdd);
        return true;
    }

    function hashCompare(bytes32 _blockhash,bytes32 _chunkhash, uint256 _diff) public pure returns(uint256 _result){
        
        uint bval = uint(_blockhash) % (2**_diff);
        uint cval = uint(_chunkhash) % (2**_diff);
        if (uint(_blockhash) == 0 || bval != cval){
            return 0;
        }else{
            return 1;            
        } 
    }

    function claimReward(uint _blockNumber, bytes32 _chunkHash, address _buyer) public returns(bool){ 
        
        if(claims[_chunkHash].claimed >= maxClaimPerChunk || _buyer == 0x0 ) return(false);
        bytes32 blockRootHash = getBlockHash(_blockNumber);
        uint chunkresult =  hashCompare(blockRootHash, _chunkHash, rewardDifficulty);
        
        if(chunkresult > 0){
            if(true){
                uint storageReward = rewardDifficulty.mul(10 ** 17);
                if (claims[_chunkHash].buyer == 0x0 ){
                    chunkList[_blockNumber].chunks.push(_chunkHash);
                    if(chunkList[_blockNumber].blockH == 0x0){
                        chunkList[_blockNumber].blockH = blockRootHash;
                        claims[_chunkHash].index = 0;
                    }else{
                        claims[_chunkHash].index = uint8(chunkList[_blockNumber].chunks.length.sub(1));
                    }
                    claims[_chunkHash].declaredBlockN = _blockNumber;
                    claims[_chunkHash].reward = storageReward;
                    claims[_chunkHash].buyer = _buyer;
                    
                    
                }
                if (!addFarmer(_chunkHash, msg.sender)) return false;
                //chunkClaimable[msg.sender] = chunkClaimable[msg.sender].add(storageReward);
                Claim(_buyer, msg.sender, _chunkHash, _blockNumber);
                claims[_chunkHash].claimed = uint8(claims[_chunkHash].farmerList.length);
                return(true);
            }else{
                //Invalid Proof of Dilivery Claims - punishment can be carried out here 
                return(false);                
            }
        }else{
            //Invalid Proof of Custody
            return(false);            
        }
    }
    
    function processReward(bytes32 _chunkToProcess, uint8[] _rejectedClaims) onlyValidators public returns(bool){
        
        var processedChunk = claims[_chunkToProcess];
        var settlingBlockN = processedChunk.declaredBlockN.add(256);
        
        address buyer = processedChunk.buyer;
        uint chunkReward = processedChunk.reward;
        uint totalCharge = (processedChunk.farmerList.length.sub(_rejectedClaims.length)).mul(chunkReward);
        // TODO: check buyer balance here
        require (block.number >= settlingBlockN && settlingBlockN != 256);
        
        if (_rejectedClaims[0] != 137){
            //nuke the rejected Claims
            for (uint i = 0; i< _rejectedClaims.length; i++){
                if (_rejectedClaims[i] < processedChunk.farmerList.length){
                    delete processedChunk.farmerList[_rejectedClaims[i]];
                }
            }
        }
        
        for (uint j = 0; j < processedChunk.farmerList.length; j++){
            var farmer = processedChunk.farmerList[j];
            if(farmer != 0x0){
                unpaidEarning[farmer] = unpaidEarning[farmer].add(chunkReward);
                Reward(buyer, farmer, msg.sender, _chunkToProcess, block.number);
            }else{
                Rejection(buyer, farmer, msg.sender, _chunkToProcess, block.number);
            }
        }

        unpaidCost[buyer] = unpaidCost[buyer].add(totalCharge);
        delete chunkList[processedChunk.declaredBlockN].chunks[processedChunk.index];
        delete claims[_chunkToProcess];
    }
    
    function settle(address _user) public returns (bool) {
        if (_user == 0x0) _user = msg.sender;
        rebalance(_user);
        uint256 outstandingBalance = unpaidCost[_user];
        if( outstandingBalance > 0){ 
            var availablePayment = ERC20(tokenAddress).allowance(_user, tokenReserve);
            if (availablePayment < unpaidCost[_user]) {
                unpaidCost[_user].sub(availablePayment);
                bool result = ERC20(tokenAddress).transferFrom(_user, tokenReserve, availablePayment);
                return result;
            }else{
                unpaidCost[_user].sub(outstandingBalance);
                ERC20(tokenAddress).transferFrom(_user, tokenReserve, outstandingBalance);
            }
            return true;
        }
    }
    
    function deposit(uint256 _wlkAmount) public returns (bool) {
        
        unpaidEarning[msg.sender] = unpaidEarning[msg.sender].add(_wlkAmount);
        if (ERC20(tokenAddress).allowance(msg.sender, tokenReserve) <= _wlkAmount) revert();
        ERC20(tokenAddress).transferFrom(msg.sender, tokenReserve, _wlkAmount);
        return true;
    }
    
    function withdrawal(address _farmer) public returns (bool) {
        if (_farmer == 0x0) _farmer = msg.sender;
        rebalance(_farmer);
        if (unpaidEarning[_farmer] > 0) {
            ERC20(tokenAddress).transferFrom(tokenReserve, _farmer, unpaidEarning[_farmer]);
            return true;
        }
    }
    
    function rebalance(address _user) public returns (bool _paymentRequired){
        if (_user == 0x0) _user = msg.sender;
        if (unpaidCost[_user] <= unpaidEarning[_user]){
            unpaidEarning[_user] = unpaidEarning[_user].sub(unpaidCost[_user]);
            unpaidCost[_user] = 0;
            return false;
        }else{
            unpaidCost[_user] = unpaidCost[_user].sub(unpaidEarning[_user]);
            unpaidEarning[_user] = 0;
            return true;
        }
    }
}
