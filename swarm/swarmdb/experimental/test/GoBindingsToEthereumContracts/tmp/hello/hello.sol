pragma solidity ^0.4.19;

contract mortal {
  /* Define variable owner of the type address */
  address owner;

  /* This function is executed at initilization and sets the owner of the contract */
  function mortal() public { owner = msg.sender; }

  /* Function to recover the funds on the contract */
  function kill() public { if (msg.sender == owner) selfdestruct(owner); }
}

contract greeter is mortal {
  /* Define variable gretting of the type string */
  string greeting;

  /* This runs when the contract is executed */
  function greeter(string _greeting) public {
    greeting = _greeting;
  }

  /* Main function */
  function greet() public constant returns (string) {
    return greeting;
  }
}