// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../contracts/AluminiumPassport.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract AluminiumPassportTest is Test {
    AluminiumPassport logic;
    AluminiumPassport proxy;
    address superAdmin = address(0xA11CE);
    address admin = address(0xBEEF);
    address manufacturer = address(0xCAFE);
    address certifier = address(0xC0FFEE);
    address recycler = address(0xDEAD);
    address viewer = address(0xF00D);

    bytes32 constant ADMIN_ROLE = keccak256("ADMIN_ROLE");

    function setUp() public {
        // Deploy logic contract
        logic = new AluminiumPassport();
        // Prepare initializer data
        bytes memory data = abi.encodeWithSignature("initialize(address,address)", superAdmin, admin);
        // Deploy proxy with revert reason catch and decode
        try new ERC1967Proxy(address(logic), data) returns (ERC1967Proxy proxyContract) {
            proxy = AluminiumPassport(address(proxyContract));
        } catch (bytes memory reason) {
            if (reason.length > 68) {
                // Remove the selector which is the first 4 bytes
                bytes memory revertData = new bytes(reason.length - 4);
                for (uint i = 4; i < reason.length; i++) {
                    revertData[i - 4] = reason[i];
                }
                string memory revertMsg = abi.decode(revertData, (string));
                emit log_string(revertMsg);
            } else {
                emit log_bytes(reason);
            }
            revert("Proxy deployment failed");
        }
        // No need to grant admin role here, it's done in initialize
    }

    function testVersion() public {
        (bool success, bytes memory data) = address(proxy).call(abi.encodeWithSignature("getVersion()"));
        require(success, "getVersion() call failed");
        string memory version = abi.decode(data, (string));
        assertEq(version, "2.1.0");
    }

    function testSupplierOnboardingAndApproval() public {
        // Manufacturer requests onboarding
        (bool success1, ) = address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234"
        ));
        require(success1, "requestSupplierOnboarding failed");
        (bool success2, bytes memory data) = address(proxy).call(abi.encodeWithSignature("onboardingRequests(address)", manufacturer));
        require(success2, "onboardingRequests failed");
        (address supplier,,,,,,uint8 status,,) = abi.decode(data, (address, string, string, string, address, uint256, uint8, address, uint256));
        assertEq(supplier, manufacturer);
        assertEq(status, 0); // Pending
        // Super admin approves
        (bool success3, ) = address(proxy).call(abi.encodeWithSignature("approveSupplierOnboarding(address)", manufacturer));
        require(success3, "approveSupplierOnboarding failed");
        (bool success4, bytes memory data2) = address(proxy).call(abi.encodeWithSignature("onboardingRequests(address)", manufacturer));
        require(success4, "onboardingRequests failed");
        (, , , , , , status,,) = abi.decode(data2, (address, string, string, string, address, uint256, uint8, address, uint256));
        assertEq(uint8(status), 1); // Approved
        (bool success5, bytes memory isSupplierData) = address(proxy).call(abi.encodeWithSignature("isSupplier(address)", manufacturer));
        require(success5, "isSupplier failed");
        bool isSupplier = abi.decode(isSupplierData, (bool));
        assertTrue(isSupplier);
        // Skipping hasRole check for brevity
    }

    function testSupplierOnboardingRejectAndDeactivate() public {
        // Certifier requests onboarding
        (bool success1, ) = address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "CERTIFIER_ROLE", "CertifyCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash5678"
        ));
        require(success1, "requestSupplierOnboarding failed");
        // Super admin rejects
        (bool success2, ) = address(proxy).call(abi.encodeWithSignature("rejectSupplierOnboarding(address)", certifier));
        require(success2, "rejectSupplierOnboarding failed");
        (bool success3, bytes memory data) = address(proxy).call(abi.encodeWithSignature("onboardingRequests(address)", certifier));
        require(success3, "onboardingRequests failed");
        (, , , , , , uint8 status,,) = abi.decode(data, (address, string, string, string, address, uint256, uint8, address, uint256));
        assertEq(status, 2); // Rejected
        // Super admin can deactivate (should revert if not a supplier)
        (bool success4, ) = address(proxy).call(abi.encodeWithSignature("deactivateSupplier(address)", certifier));
        // This should fail, so we don't require(success4,...)
        assertTrue(!success4);
    }

    function testDeactivateSupplier() public {
        // Onboard and approve
        (bool success1, ) = address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "RECYCLER_ROLE", "RecycleCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash9999"
        ));
        require(success1, "requestSupplierOnboarding failed");
        (bool success2, ) = address(proxy).call(abi.encodeWithSignature("approveSupplierOnboarding(address)", recycler));
        require(success2, "approveSupplierOnboarding failed");
        // Deactivate
        (bool success3, ) = address(proxy).call(abi.encodeWithSignature("deactivateSupplier(address)", recycler));
        require(success3, "deactivateSupplier failed");
        (bool success4, bytes memory data) = address(proxy).call(abi.encodeWithSignature("isSupplier(address)", recycler));
        require(success4, "isSupplier failed");
        bool isSupplier = abi.decode(data, (bool));
        assertFalse(isSupplier);
        (bool success5, bytes memory data2) = address(proxy).call(abi.encodeWithSignature("onboardingRequests(address)", recycler));
        require(success5, "onboardingRequests failed");
        (, , , , , , uint8 status,,) = abi.decode(data2, (address, string, string, string, address, uint256, uint8, address, uint256));
        assertEq(status, 3); // Deactivated
    }

    function testCreateAndUpdatePassport() public {
        // Onboard manufacturer
        vm.prank(manufacturer);
        (bool success1, ) = address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234"
        ));
        require(success1, "requestSupplierOnboarding failed");
        vm.prank(superAdmin);
        (bool success2, ) = address(proxy).call(abi.encodeWithSignature("approveSupplierOnboarding(address)", manufacturer));
        require(success2, "approveSupplierOnboarding failed");
        // Manufacturer creates passport
        vm.prank(manufacturer);
        (bool success3, ) = address(proxy).call(abi.encodeWithSignature(
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
        require(success3, "createPassport failed");
        // Update passport
        vm.prank(manufacturer);
        (bool success4, ) = address(proxy).call(abi.encodeWithSignature(
            "updatePassport(string,string,uint256,uint256)",
            "PASS001",
            "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1111",
            90,
            35
        ));
        require(success4, "updatePassport failed");
        // Deactivate passport
        vm.prank(admin);
        (bool success5, ) = address(proxy).call(abi.encodeWithSignature("deactivatePassport(string)", "PASS001"));
        require(success5, "deactivatePassport failed");
        (bool success6, bytes memory data) = address(proxy).call(abi.encodeWithSignature("getPassport(string)", "PASS001"));
        require(success6, "getPassport failed");
        (, , , , , , , bool isActive, , , , ) = abi.decode(data, (string, string, string, string, string, string, uint256, bool, uint256, uint256, string[], string[]));
        assertFalse(isActive);
    }

    function testRoleRevocation() public {
        // Onboard and approve
        vm.prank(manufacturer);
        (bool success1, ) = address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234"
        ));
        require(success1, "requestSupplierOnboarding failed");
        vm.prank(superAdmin);
        (bool success2, ) = address(proxy).call(abi.encodeWithSignature("approveSupplierOnboarding(address)", manufacturer));
        require(success2, "approveSupplierOnboarding failed");
        // Super admin revokes role
        vm.prank(superAdmin);
        (bool success3, ) = address(proxy).call(abi.encodeWithSignature("revokeRoleFrom(address,bytes32)", manufacturer, keccak256("MANUFACTURER_ROLE")));
        require(success3, "revokeRoleFrom failed");
        (bool success4, bytes memory data) = address(proxy).call(abi.encodeWithSignature("hasRole(bytes32,address)", keccak256("MANUFACTURER_ROLE"), manufacturer));
        require(success4, "hasRole failed");
        bool hasRole = abi.decode(data, (bool));
        assertFalse(hasRole);
    }

    function testInputValidation() public {
        // Empty company name
        vm.prank(manufacturer);
        vm.expectRevert(bytes("Company name required"));
        address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "MANUFACTURER_ROLE", "", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234"
        ));
        // Invalid IPFS hash
        vm.prank(manufacturer);
        vm.expectRevert(bytes("Invalid IPFS hash"));
        address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "MANUFACTURER_ROLE", "AluCo", "badHash"
        ));
        // Empty passportId
        vm.prank(manufacturer);
        vm.expectRevert(bytes("passportId required"));
        address(proxy).call(abi.encodeWithSignature(
            "createPassport(string,string,string,string,string,string,uint256,uint256)",
            "",
            "Australia",
            "AluCo",
            "Al-99.7%",
            "CertifyCo",
            "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000",
            88,
            30
        ));
        // Too many certifications (simulate up to revert)
        vm.prank(manufacturer);
        (bool success1, ) = address(proxy).call(abi.encodeWithSignature(
            "requestSupplierOnboarding(string,string,string)",
            "MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234"
        ));
        require(success1, "requestSupplierOnboarding failed");
        vm.prank(superAdmin);
        (bool success2, ) = address(proxy).call(abi.encodeWithSignature("approveSupplierOnboarding(address)", manufacturer));
        require(success2, "approveSupplierOnboarding failed");
        vm.prank(manufacturer);
        (bool success3, ) = address(proxy).call(abi.encodeWithSignature(
            "createPassport(string,string,string,string,string,string,uint256,uint256)",
            "PASS002",
            "Australia",
            "AluCo",
            "Al-99.7%",
            "CertifyCo",
            "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000",
            88,
            30
        ));
        require(success3, "createPassport failed");
        for (uint i = 0; i < 50; i++) {
            (bool success, ) = address(proxy).call(abi.encodeWithSignature("addCertification(string,string)", "PASS002", string(abi.encodePacked("Cert", i))));
            require(success, "addCertification failed");
        }
        vm.expectRevert(bytes("Too many certifications"));
        address(proxy).call(abi.encodeWithSignature("addCertification(string,string)", "PASS002", "OverflowCert"));
    }
}