import "mortal";

/// @title Chequebook for Ethereum micropayments
/// @author Daniel A. Nagy <daniel@ethdev.com>
contract chequebook is mortal {
    // Cumulative paid amount in wei to each beneficiary
    mapping (address => uint256) sent;

    /// @notice Overdraft event
    event Overdraft(address deadbeat);
    
    /// @notice Cash cheque
    /// 
    /// @param beneficiary beneficiary address
    /// @param amount cumulative amount in wei
    /// @param sig_v signature parameter v
    /// @param sig_r signature parameter r
    /// @param sig_s signature parameter s
    function cash(address beneficiary, uint256 amount,
        uint8 sig_v, bytes32 sig_r, bytes32 sig_s) {
        if(amount <= sent[beneficiary]) return;
        bytes32 hash = sha3(beneficiary, amount);
        if(owner != ecrecover(hash, sig_v, sig_r, sig_s)) return;
        if (beneficiary.send(amount - sent[beneficiary])) {
            sent[beneficiary] = amount;
        } else {
            // owner.sendToDebtorsPrison();
            Overdraft(owner);
            suicide(beneficiary);
        }
    }
}

