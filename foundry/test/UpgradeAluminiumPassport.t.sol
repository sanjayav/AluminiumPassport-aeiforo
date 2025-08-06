// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../contracts/AluminiumPassport.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

/// @notice Test the upgrade process for AluminiumPassport UUPS proxy
contract UpgradeAluminiumPassportTest is Test {
    AluminiumPassport logicV1;
    AluminiumPassport proxy;
    address superAdmin = address(0xA11CE);
    address admin = address(0xBEEF);

    function setUp() public {
        // Deploy initial logic contract (V1)
        logicV1 = new AluminiumPassport();
        // Prepare initializer data
        bytes memory data = abi.encodeWithSignature("initialize(address,address)", superAdmin, admin);
        // Deploy proxy
        ERC1967Proxy proxyContract = new ERC1967Proxy(address(logicV1), data);
        proxy = AluminiumPassport(address(proxyContract));
        // Grant admin role (already done in initialize)
    }

    function testUpgradeProcess() public {
        // Check initial version
        (bool success, bytes memory data) = address(proxy).call(abi.encodeWithSignature("getVersion()"));
        require(success, "getVersion() call failed");
        string memory version = abi.decode(data, (string));
        assertEq(version, "2.1.0");
        // Deploy new logic contract (V2, could be the same for this test)
        AluminiumPassport logicV2 = new AluminiumPassport();
        // Upgrade proxy to new logic as super admin
        vm.prank(superAdmin);
        (bool upgradeSuccess, ) = address(proxy).call(abi.encodeWithSignature("upgradeTo(address)", address(logicV2)));
        require(upgradeSuccess, "upgradeTo failed");
        // Check that version is still correct (since logic is the same)
        (success, data) = address(proxy).call(abi.encodeWithSignature("getVersion()"));
        require(success, "getVersion() call failed");
        version = abi.decode(data, (string));
        assertEq(version, "2.1.0");
        // Check that roles are preserved
        (bool hasSuperAdminRoleSuccess, bytes memory hasSuperAdminRoleData) = address(proxy).call(abi.encodeWithSignature("hasRole(bytes32,address)", keccak256("SUPER_ADMIN_ROLE"), superAdmin));
        require(hasSuperAdminRoleSuccess, "hasRole call failed");
        bool hasSuperAdminRole = abi.decode(hasSuperAdminRoleData, (bool));
        assertTrue(hasSuperAdminRole);
        (bool hasAdminRoleSuccess, bytes memory hasAdminRoleData) = address(proxy).call(abi.encodeWithSignature("hasRole(bytes32,address)", keccak256("ADMIN_ROLE"), admin));
        require(hasAdminRoleSuccess, "hasRole call failed");
        bool hasAdminRole = abi.decode(hasAdminRoleData, (bool));
        assertTrue(hasAdminRole);
        // Check that new logic features work (pause/unpause)
        vm.prank(superAdmin);
        (bool pauseSuccess, ) = address(proxy).call(abi.encodeWithSignature("pause()"));
        require(pauseSuccess, "pause failed");
        (bool isPausedSuccess, bytes memory isPausedData) = address(proxy).call(abi.encodeWithSignature("isPaused()"));
        require(isPausedSuccess, "isPaused call failed");
        bool isPaused = abi.decode(isPausedData, (bool));
        assertTrue(isPaused);
        vm.prank(superAdmin);
        (bool unpauseSuccess, ) = address(proxy).call(abi.encodeWithSignature("unpause()"));
        require(unpauseSuccess, "unpause failed");
        (isPausedSuccess, isPausedData) = address(proxy).call(abi.encodeWithSignature("isPaused()"));
        require(isPausedSuccess, "isPaused call failed");
        isPaused = abi.decode(isPausedData, (bool));
        assertFalse(isPaused);
    }

    function testUpgradePreservesState() public {
        // Set some state before upgrade
        vm.prank(admin);
        (bool createSuccess, ) = address(proxy).call(abi.encodeWithSignature(
            "createPassport(string,string,string,string,string,string,uint256,uint256)",
            "PASS001",
            "Australia",
            "AluCo",
            "Al-99.7%",
            "CertifyCo",
            "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000",
            88,
            30
        ));
        require(createSuccess, "createPassport failed");
        // Deploy new logic contract
        AluminiumPassport logicV2 = new AluminiumPassport();
        // Upgrade
        vm.prank(superAdmin);
        (bool upgradeSuccess, ) = address(proxy).call(abi.encodeWithSignature("upgradeTo(address)", address(logicV2)));
        require(upgradeSuccess, "upgradeTo failed");
        // Check that passport data is still there
        (bool getSuccess, bytes memory data) = address(proxy).call(abi.encodeWithSignature("getPassport(string)", "PASS001"));
        require(getSuccess, "getPassport failed");
        (string memory pid, string memory origin, string memory manufacturer, string memory alloyComposition, string memory certifier, uint256 esgScore, uint256 recycledContent, bool isActive, string memory ipfsHash, address createdBy, uint256 createdAt, uint256 updatedAt) = abi.decode(data, (string, string, string, string, string, uint256, uint256, bool, string, address, uint256, uint256));
        assertEq(pid, "PASS001");
        assertEq(recycledContent, 30);
        assertEq(ipfsHash, "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000");
    }
}