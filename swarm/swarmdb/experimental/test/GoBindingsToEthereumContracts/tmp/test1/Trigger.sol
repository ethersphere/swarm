contract Trigger {
  function () {
      throw;
  }

  address owner;

  function Trigger() {
      owner = msg.sender;
  }

  event TriggerEvt(address _sender, uint _trigger);

  function trigger(uint _trigger) {
      TriggerEvt(msg.sender, _trigger);
  }

  function getOwner() constant returns (address) {
    return owner;
  }

}