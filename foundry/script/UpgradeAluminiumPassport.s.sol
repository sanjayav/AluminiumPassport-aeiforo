// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../contracts/AluminiumPassport.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

/// @notice Foundry script to upgrade the AluminiumPassport proxy to a new implementation
contract UpgradeAluminiumPassport is Script {
    /// @notice Deploys new logic and upgrades the proxy
    function run() external {
        address superAdmin = vm.envAddress("SUPER_ADMIN");
        address proxyAddr = vm.envAddress("PROXY_ADDRESS");
        vm.startBroadcast(superAdmin);

        // Deploy new logic contract
        AluminiumPassport newLogic = new AluminiumPassport();
        console.log("New logic contract deployed at:", address(newLogic));

        // Call upgradeTo on the proxy as super admin
        // The proxy must already be a UUPS proxy pointing to the old implementation
        // We use the proxy address as an AluminiumPassport instance
        AluminiumPassport proxy = AluminiumPassport(proxyAddr);
        proxy.upgradeTo(address(newLogic));
        console.log("Proxy upgraded at:", proxyAddr);
        console.log("Proxy now points to logic:", address(newLogic));

        vm.stopBroadcast();
    }
}