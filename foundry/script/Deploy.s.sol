
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Script.sol";
import "../contracts/AluminiumPassport.sol";

contract DeployPassport is Script {
    function setUp() public {}

    function run() public {
        vm.startBroadcast();
        new AluminiumPassport();
        vm.stopBroadcast();
    }
}
