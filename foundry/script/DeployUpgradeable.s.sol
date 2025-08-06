// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../contracts/AluminiumPassport.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

/// @notice Foundry script to deploy AluminiumPassport as a UUPS upgradeable proxy
contract DeployUpgradeable is Script {
    function run() external {
        // Get the super admin address from environment variable
        address superAdmin = vm.envAddress("SUPER_ADMIN");
        vm.startBroadcast();

        // Deploy the logic (implementation) contract
        AluminiumPassport logic = new AluminiumPassport();

        // Prepare initializer data for proxy
        bytes memory data = abi.encodeWithSignature("initialize(address)", superAdmin);

        // Deploy the ERC1967Proxy (UUPS proxy)
        ERC1967Proxy proxy = new ERC1967Proxy(address(logic), data);

        // Print deployed addresses
        console.log("Logic contract address:", address(logic));
        console.log("Proxy contract address:", address(proxy));

        vm.stopBroadcast();
    }
}