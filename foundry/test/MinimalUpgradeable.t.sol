// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../../contracts/MinimalUpgradeable.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract MinimalUpgradeableTest is Test {
    MinimalUpgradeable logic;
    MinimalUpgradeable proxy;

    function setUp() public {
        logic = new MinimalUpgradeable();
        bytes memory data = abi.encodeWithSignature("initialize(uint256)", 42);
        ERC1967Proxy proxyContract = new ERC1967Proxy(address(logic), data);
        proxy = MinimalUpgradeable(address(proxyContract));
    }

    function testValue() public {
        assertEq(proxy.value(), 42);
    }
}