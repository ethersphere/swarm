pragma solidity ^0.4.18;

// Translation in progress of https://github.com/ethereum/casper/blob/master/casper/contracts/simple_casper.v.py
// Skipped: sig checks

import "github.com/wolkdb/swarm.wolk.com/src/github.com/ethereum/go-ethereum/swarmdb/contracts/plasma/RLP.sol";

contract Casper {
 
    using RLP for bytes;
    using RLP for RLP.RLPItem;
    using RLP for RLP.Iterator;
    
    // Information about validators
    struct Validator {
        // Used to determine the amount of wei the validator holds. To get the actual amount of wei, multiply this by the deposit_scale_factor.
        uint256 deposit;
        // The dynasty the validator is joining
        uint256 start_dynasty;
        // The dynasty the validator is leaving
        uint256 end_dynasty;
        // The address which the validator's signatures must verify to (to be later replaced with validation code)
        address addr;
        // Address to withdraw to
        address withdrawal_addr;
    }

    mapping (uint => Validator) validators; 
    
    // Historical checkoint hashes
    mapping (uint => bytes32) checkpoint_hashes;
    
    // Number of validators
    uint public nextValidatorIndex;
    
    // Mapping of validator's signature address to their index number
    mapping (address => uint) validator_indexes;
    
    // The current dynasty (validator set changes between dynasties)
    uint public dynasty;

    // Amount of wei added to the total deposits in the next dynasty
    uint256 public next_dynasty_wei_delta; 
  
    // Amount of wei added to the total deposits in the dynasty after that
    uint256 public second_next_dynasty_wei_delta;

    // Total deposits in the current dynasty
    uint256 public total_curdyn_deposits;

    // Total deposits in the previous dynasty
    uint256 public total_prevdyn_deposits;
    
    // Mapping of dynasty to start epoch of that dynasty
    mapping (uint => uint) public dynasty_start_epoch;

    // Mapping of epoch to what dynasty it is
    mapping (uint => uint) public dynasty_in_epoch;
      
    // Information for use in processing cryptoeconomic commitments
    struct Vote {
        // How many votes are there for this source epoch from the current dynasty
        mapping (uint => uint256) cur_dyn_votes;
        // From the previous dynasty
        mapping (uint => uint256) prev_dyn_votes;
        // Bitmap of which validator IDs have already voted
        mapping (uint => mapping (bytes32 => uint256)) vote_bitmap;
        // Is a vote referencing the given epoch justified?
        bool is_justified;
        // Is a vote referencing the given epoch finalized?
        bool is_finalized;
    }
    mapping (uint => Vote) public votes; 
    // index: target epoch

    // Is the current expected hash justified
    bool public main_hash_justified;

    // Value used to calculate the per-epoch fee that validators should be charged
    mapping (uint => uint256) public deposit_scale_factor;

    // For debug purposes: Remove this when ready.
    uint256 public last_nonvoter_rescale;
    uint256 public last_voter_rescale;
      
    // Length of an epoch in blocks
    uint public epoch_length;

    // Withdrawal delay in blocks
    uint public withdrawal_delay;

    // Current epoch
    uint public current_epoch;
      
    // Last finalized epoch
    uint public last_finalized_epoch;

    // Last justified epoch
    uint public last_justified_epoch;

    // Expected source epoch for a vote
    uint public expected_source_epoch;

    // Total deposits destroyed
    uint256 public total_destroyed;
    
    // Sighash calculator library address
    address sighasher;

    // Purity checker library address
    address purity_checker;
    
    // Reward for voting as fraction of deposit size
    uint256 public reward_factor;

    // Base interest factor
    uint256 base_interest_factor;

    // Base penalty factor
    uint256 base_penalty_factor;

    // Log topic for vote
    bytes32 vote_log_topic;

    // Minimum deposit size if no one else is validating
    uint256 min_deposit_size;

    address owner; // could use Owned pattern
    function init(  
                  // Epoch length, delay in epochs for withdrawing
                  uint256 _epoch_length,
                  uint256 _withdrawal_delay,
                  // Owner (backdoor), sig hash calculator, purity checker
                  address _owner, 
                  address _sighasher, 
                  address _purity_checker,
                  // Base interest and base penalty factors
                  uint256 _base_interest_factor,
                  uint256 _base_penalty_factor,
                  // Min deposit size (in wei)
                  uint256 _min_deposit_size) public {
      // Epoch length
      epoch_length = _epoch_length;
      // Delay in epochs for withdrawing
      withdrawal_delay = _withdrawal_delay;
      // Start validator index counter at 1 because validator_indexes[] requires non-zero values
      nextValidatorIndex = 1;
      // Temporary backdoor for testing purposes (to allow recovering destroyed deposits)
      owner = _owner;
      // Set deposit scale factor
      deposit_scale_factor[0] = 10000000000;
      // Start dynasty counter at 0
      dynasty = 0;
      // Initialize the epoch counter
      current_epoch = block.number / epoch_length;
      // Set the sighash calculator address
      sighasher = _sighasher;
      // Set the purity checker address
      purity_checker = _purity_checker;
      // votes[0].committed = True
      // Set initial total deposit counter
      total_curdyn_deposits = 0;
      total_prevdyn_deposits = 0;
      // Constants that affect interest rates and penalties
      base_interest_factor = _base_interest_factor;
      base_penalty_factor = _base_penalty_factor;
      vote_log_topic = keccak256("vote()");  // was: sha3
      // Constants that affect the min deposit size
      min_deposit_size = _min_deposit_size;
    }
    
    function min(uint x, uint y) pure private returns(uint) {
            if (x<y) {
                return(x);
            }
            return(y);
    }
    // TODO: fix decimal => uint256 here and for all deposit_scale_factor
    function get_main_hash_voted_frac() public constant returns(uint256) {
      return min(votes[current_epoch].cur_dyn_votes[expected_source_epoch] / total_curdyn_deposits,
                 votes[current_epoch].prev_dyn_votes[expected_source_epoch] / total_prevdyn_deposits);
    }
    
    function get_deposit_size(uint validator_index)  public constant returns(uint256) {
      return (validators[validator_index].deposit * deposit_scale_factor[current_epoch]);
    }

    function get_total_curdyn_deposits() public constant returns(uint256) {
      return (total_curdyn_deposits * deposit_scale_factor[current_epoch]);
    }

    function get_total_prevdyn_deposits() public constant returns (uint256) {
      return (total_prevdyn_deposits * deposit_scale_factor[current_epoch]);
    }
    
    // Helper functions that clients can call to know what to vote
    function get_recommended_source_epoch() public constant returns (uint256) {
      return expected_source_epoch;
    }

    function get_recommended_target_hash() public constant returns (bytes32) {
      return block.blockhash(current_epoch*epoch_length-1);
    }
    
    function deposit_exists() private constant returns(bool) {
      return(total_curdyn_deposits > 0 && total_prevdyn_deposits > 0);
    }
    
    // ***** Private *****
    // Increment dynasty when checkpoint is finalized. Might want to split out the cases separately.
    function increment_dynasty() private {
        uint epoch = current_epoch;
        // Increment the dynasty if finalized
        if ( votes[epoch-2].is_finalized ) {
            dynasty += 1;
            total_prevdyn_deposits = total_curdyn_deposits;
            total_curdyn_deposits += next_dynasty_wei_delta;
            next_dynasty_wei_delta = second_next_dynasty_wei_delta;
            second_next_dynasty_wei_delta = 0;
            dynasty_start_epoch[dynasty] = epoch;
        }
        dynasty_in_epoch[epoch] = dynasty;
        if ( main_hash_justified ) {
            expected_source_epoch = epoch - 1;
        }
        main_hash_justified = false;
    }

    // Returns number of epochs since finalization.
    function get_esf() private view returns (uint) {
        uint epoch = current_epoch;
        return(epoch - last_finalized_epoch);
    }
      
    // Returns the current collective reward factor, which rewards the dynasty for high-voting levels.
    function get_collective_reward() private view returns (uint256) {
        uint epoch = current_epoch;
        bool live = ( get_esf() <= 2 );
        if ( ! deposit_exists() || ! live ) {
            return(0);
        }
        // TODO: Fraction that voted
        uint256 cur_vote_frac = votes[epoch - 1].cur_dyn_votes[expected_source_epoch] / total_curdyn_deposits;
        uint256 prev_vote_frac = votes[epoch - 1].prev_dyn_votes[expected_source_epoch] / total_prevdyn_deposits;
        uint256 vote_frac = min(cur_vote_frac, prev_vote_frac);
        return(vote_frac * reward_factor / 2);
    }

    function insta_finalize() private {
        uint epoch = current_epoch;
        main_hash_justified = true;
        votes[epoch - 1].is_justified = true;
        votes[epoch - 1].is_finalized = true;
        last_justified_epoch = epoch - 1;
        last_finalized_epoch = epoch - 1;
    }

    function get_sqrt_of_total_deposits() private pure returns (uint256) {
        // uint epoch = current_epoch;
        uint256 ether_deposited_as_number =  1; // TODO: (max(total_prevdyn_deposits, total_curdyn_deposits) * deposit_scale_factor[epoch - 1] / as_wei_value(1, "ether")) + 1;
        uint256 sqrt = ether_deposited_as_number / 2.0;
        for (uint i = 0; i<20; i++) {
            sqrt = (sqrt + (ether_deposited_as_number / sqrt)) / 2;
        }
        return sqrt;
    }

    // Called at the start of any epoch
    function initialize_epoch(uint epoch) private {
        // Check that the epoch actually has started
        uint computed_current_epoch = block.number / epoch_length;
        assert(epoch <= computed_current_epoch && epoch == current_epoch + 1);
      
        // Setup
        current_epoch = epoch;
      
        // Reward if finalized at least in the last two epochs
        last_nonvoter_rescale = (1 + get_collective_reward() - reward_factor);
        last_voter_rescale = last_nonvoter_rescale * (1 + reward_factor);
        deposit_scale_factor[epoch] = deposit_scale_factor[epoch - 1] * last_nonvoter_rescale;
      
        if ( deposit_exists() ) {
            // Set the reward factor for the next epoch.
            uint256 adj_interest_base = base_interest_factor / get_sqrt_of_total_deposits();
            // sqrt is based on previous epoch starting deposit
            reward_factor = adj_interest_base + base_penalty_factor * get_esf();  // might not be bpf. clarify is positive?
            // ESF is only thing that is changing and reward_factor is being used above.
            assert(reward_factor > 0);
        } else {
            insta_finalize();  // comment on why.
            reward_factor = 0;
        }
      
        // Increment the dynasty if finalized
        increment_dynasty();
      
        // Store checkpoint hash for easy access
        checkpoint_hashes[epoch] = get_recommended_target_hash();
    }

    // Send a deposit to join the validator set
    function deposit(address validation_addr, address withdrawal_addr) public payable {
        assert(current_epoch == block.number / epoch_length);
        // TODO: assert(extract32(raw_call(purity_checker, concat('\xa1\x90>\xab', as_bytes32(validation_addr)), gas=500000, outsize=32), 0) != as_bytes32(0));
        // assert(! validator_indexes[withdrawal_addr]);
        assert(msg.value >= min_deposit_size);
        validators[nextValidatorIndex] = Validator({deposit: msg.value / deposit_scale_factor[current_epoch], 
                             start_dynasty: dynasty + 2, 
                             end_dynasty: 1000000000000000000000000000000,
                             addr: validation_addr,
                             withdrawal_addr: withdrawal_addr});
      
        validator_indexes[withdrawal_addr] = nextValidatorIndex;
        nextValidatorIndex += 1;
        second_next_dynasty_wei_delta += msg.value / deposit_scale_factor[current_epoch];
    }

    // Log in or log out from the validator set. A logged out validator can log back in later, 
    // if they do not log in for an entire withdrawal period, they can get their money out
    function logout(bytes logout_msg) public {
        assert( current_epoch == block.number / epoch_length);
        // Get hash for signature, and implicitly assert that it is an RLP list consisting solely of RLP elements
        // bytes32 sighash; // TODO: extract32(raw_call(sighasher, logout_msg, gas=200000, outsize=32), 0);
        // Extract parameters
        var values = logout_msg.toRLPItem().toList(); // [num, num, bytes]
        uint validator_index = values[0].toUint();
        uint epoch = values[1].toUint();
        // bytes memory sig = values[2].toBytes();
        assert( current_epoch >= epoch);
        // Signature check
        // TODO: assert( extract32(raw_call(validators[validator_index].addr, concat(sighash, sig), gas=500000, outsize=32), 0) == as_bytes32(1) );
        // Check that we haven't already withdrawn
        assert(validators[validator_index].end_dynasty > dynasty + 2);
        // Set the end dynasty
        validators[validator_index].end_dynasty = dynasty + 2;
        second_next_dynasty_wei_delta -= validators[validator_index].deposit;
    }

    // Removes a validator from the validator pool
    function delete_validator(uint validator_index) public {
        if ( validators[validator_index].end_dynasty > dynasty + 2 ) {
            next_dynasty_wei_delta -= validators[validator_index].deposit;
        }
        validator_indexes[validators[validator_index].withdrawal_addr] = 0;
        validators[validator_index] = Validator( {deposit: 0, start_dynasty: 0, end_dynasty: 0, addr: 0x0, withdrawal_addr: 0x0});
    }

    // Withdraw deposited ether
    function withdraw(uint validator_index) public {
      // Check that we can withdraw
      assert(dynasty >= validators[validator_index].end_dynasty + 1);
      uint end_epoch = dynasty_start_epoch[validators[validator_index].end_dynasty + 1];
      assert(current_epoch >= end_epoch + withdrawal_delay);
      // TODO: floor/decimal
      uint256 withdraw_amount = (validators[validator_index].deposit * deposit_scale_factor[end_epoch]);
      validators[validator_index].withdrawal_addr.transfer(withdraw_amount);
      delete_validator(validator_index);
    }

    // Reward the given validator & miner, and reflect this in total deposit figured
    function proc_reward(uint validator_index, uint256 reward) private {
      // uint start_epoch = dynasty_start_epoch[validators[validator_index].start_dynasty];
      validators[validator_index].deposit += reward;
      uint start_dynasty = validators[validator_index].start_dynasty;
      uint end_dynasty = validators[validator_index].end_dynasty;
      uint current_dynasty = dynasty;
      uint past_dynasty = current_dynasty - 1;
      if ((start_dynasty <= current_dynasty) && (current_dynasty < end_dynasty)) {
        total_curdyn_deposits += reward;
      }
      if ((start_dynasty <= past_dynasty) && (past_dynasty < end_dynasty)) {
        total_prevdyn_deposits += reward;
      }
      if ( current_dynasty == end_dynasty - 1 ) {
        next_dynasty_wei_delta -= reward;
      }
      if ( current_dynasty == end_dynasty - 2 ) {
        second_next_dynasty_wei_delta -= reward;
      }
      block.coinbase.transfer((reward * deposit_scale_factor[current_epoch] / 8));
    }

    // Process a vote message
    function vote(bytes vote_msg) public {
        // Get hash for signature, and implicitly assert that it is an RLP list consisting solely of RLP elements
        //bytes32 sighash; // TODO: extract32(raw_call(sighasher, vote_msg, gas=200000, outsize=32), 0);
        // Extract parameters
        var values = vote_msg.toRLPItem().toList(); // [num, bytes32, num, num, bytes]
        uint validator_index = values[0].toUint();
        bytes32 target_hash = values[1].toBytes32();
        uint target_epoch = values[2].toUint();
        uint source_epoch = values[3].toUint();
        // bytes memory sig = values[4].toBytes();

        // Check the signature
        // TODO: assert(extract32(raw_call(validators[validator_index].addr, concat(sighash, sig), gas=500000, outsize=32), 0) == as_bytes32(1));

        // Check that this vote has not yet been made
        assert( ( votes[target_epoch].vote_bitmap[validator_index / 256][target_hash] & ( ( validator_index % 256) << 1 ) ) == 0 );
        // Check that the vote's target hash is correct
        assert(target_hash == get_recommended_target_hash());
        // Check that the vote source points to a justified epoch
        assert(votes[source_epoch].is_justified);
        vote0(validator_index, target_hash, target_epoch, source_epoch);
    }

    function record_vote(uint validator_index, bytes32 target_hash, uint target_epoch) private {
        votes[target_epoch].vote_bitmap[validator_index / 256][target_hash] = (votes[target_epoch].vote_bitmap[validator_index / 256][target_hash] | ( validator_index % 256 << 1 ) );
    }
    
    function vote0(uint validator_index, bytes32 target_hash, uint target_epoch, uint source_epoch) private {
        // Check that we are at least (epoch length / 4) blocks into the epoch assert block.number % epoch_length >= epoch_length / 4
        // Original starting dynasty of the validator; fail if before 
        uint start_dynasty = validators[validator_index].start_dynasty;
        // Ending dynasty of the current login period
        uint end_dynasty = validators[validator_index].end_dynasty;
        // Dynasty of the vote
        uint current_dynasty = dynasty_in_epoch[target_epoch];
        uint past_dynasty = current_dynasty - 1;
        bool in_current_dynasty = ((start_dynasty <= current_dynasty) && (current_dynasty < end_dynasty));
        bool in_prev_dynasty = ((start_dynasty <= past_dynasty) && (past_dynasty < end_dynasty));
        assert( in_current_dynasty || in_prev_dynasty);
        // Record that the validator voted for this target epoch so they can't again
        record_vote(validator_index, target_hash, target_epoch);
      
        // Record that this vote took place
        uint256 current_dynasty_votes = votes[target_epoch].cur_dyn_votes[source_epoch];
        uint256 previous_dynasty_votes = votes[target_epoch].prev_dyn_votes[source_epoch];
        if ( in_current_dynasty ) {
            current_dynasty_votes += validators[validator_index].deposit;
            votes[target_epoch].cur_dyn_votes[source_epoch] = current_dynasty_votes;
        } 
        if ( in_prev_dynasty ) {
            previous_dynasty_votes += validators[validator_index].deposit;
            votes[target_epoch].prev_dyn_votes[source_epoch] = previous_dynasty_votes;
        }

        // Process rewards. Check that we have not yet voted for this target_epoch
        // Pay the reward if the vote was submitted in time and the vote is voting the correct data
        if ( current_epoch == target_epoch && expected_source_epoch == source_epoch ) {
            uint256 reward = (validators[validator_index].deposit * reward_factor);
            proc_reward(validator_index, reward);
        }

        // If enough votes with the same source_epoch and hash are made, then the hash value is justified
        if ( ( current_dynasty_votes >= total_curdyn_deposits * 2 / 3 ) && ( previous_dynasty_votes >= total_prevdyn_deposits * 2 / 3 ) && ( ! votes[target_epoch].is_justified ) ) {
            votes[target_epoch].is_justified = true;
            last_justified_epoch = target_epoch;
            if ( target_epoch == current_epoch ) {
                main_hash_justified = true;
            }
            // If two epochs are justified consecutively,
            // then the source_epoch finalized
            if ( target_epoch == source_epoch + 1 ) {
                votes[source_epoch].is_finalized = true;
                last_finalized_epoch = source_epoch;
            }
            // raw_log([vote_log_topic], vote_msg);
        }
    }

    // Cannot make two prepares in the same epoch; no surround vote.
    function slash(bytes vote_msg_1, bytes vote_msg_2) public {
      // Message 1: Extract parameters [num, bytes32, num, num, bytes]
      var values_1 = vote_msg_1.toRLPItem().toList(); 
      uint validator_index_1 = values_1[0].toUint();
      uint target_epoch_1 = values_1[2].toUint();
      uint source_epoch_1 = values_1[3].toUint();
      // bytes memory sig_1 = values_1[4].toBytes();
      // bytes32 sighash_1 = internalhash(vote_msg_1, 128); 
      // Check the signature for vote message 1
      // TODO: assert(extract32(raw_call(validators[validator_index_1].addr, concat(sighash_1, sig_1), gas=500000, outsize=32), 0) == as_bytes32(1));
      // Message 2: Extract parameters (Same as Message 1)
      // bytes memory sighash_2; // TODO: extract32(raw_call(sighasher, vote_msg_2, gas=200000, outsize=32), 0);
      var values_2 = vote_msg_2.toRLPItem().toList(); // [num, bytes32, num, num, bytes]);
      uint validator_index_2 = values_2[0].toUint();
      uint target_epoch_2 = values_2[2].toUint();
      uint source_epoch_2 = values_2[3].toUint();
      // bytes memory sig_2 = values_2[4].toBytes();
      // Check the signature for vote message 2
      // assert(extract32(raw_call(validators[validator_index_2].addr, concat(sighash_2, sig_2), gas=500000, outsize=32), 0) == as_bytes32(1));
      // Check the messages are from the same validator
      assert(validator_index_1 == validator_index_2);
      // Check the messages are not the same
      // TODO: assert(sighash_1 != sighash_2);
      // Detect slashing
      bool slashing_condition_detected = false;
      if ( target_epoch_1 == target_epoch_2 ) { // NO DBL VOTE
        slashing_condition_detected = true;
      } else if ( (target_epoch_1 > target_epoch_2 && source_epoch_1 < source_epoch_2) || (target_epoch_2 > target_epoch_1  &&  source_epoch_2 < source_epoch_1) ) {
        // NO SURROUND VOTE
        slashing_condition_detected = true;
      }

      assert(slashing_condition_detected);
      // Delete the offending validator, and give a 4% "finder's fee"
      uint256 validator_deposit = get_deposit_size(validator_index_1);
      uint256 slashing_bounty = validator_deposit / 25;
      total_destroyed += validator_deposit * 24 / 25;
      delete_validator(validator_index_1);
      msg.sender.transfer(slashing_bounty);
    }
}
