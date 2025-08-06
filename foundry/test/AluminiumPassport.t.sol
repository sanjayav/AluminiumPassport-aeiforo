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

    function setUp() public {
        // Deploy logic contract
        logic = new AluminiumPassport();
        // Prepare initializer data
        bytes memory data = abi.encodeWithSelector(
            AluminiumPassport.initialize.selector,
            superAdmin
        );
        // Deploy proxy
        ERC1967Proxy proxyContract = new ERC1967Proxy(address(logic), data);
        proxy = AluminiumPassport(address(proxyContract));
        // Grant roles
        vm.prank(superAdmin);
        proxy.grantRole(proxy.ADMIN_ROLE(), admin);
    }

    function testVersion() public {
        assertEq(proxy.getVersion(), "2.0.0");
    }

    function testSupplierOnboardingAndApproval() public {
        // Manufacturer requests onboarding
        vm.prank(manufacturer);
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
        (address supplier,,,,,,AluminiumPassport.OnboardingStatus status,,) = proxy.onboardingRequests(manufacturer);
        assertEq(supplier, manufacturer);
        assertEq(uint256(status), uint256(AluminiumPassport.OnboardingStatus.Pending));
        // Super admin approves
        vm.prank(superAdmin);
        proxy.approveSupplierOnboarding(manufacturer);
        (, , , , , , status,,) = proxy.onboardingRequests(manufacturer);
        assertEq(uint256(status), uint256(AluminiumPassport.OnboardingStatus.Approved));
        assertTrue(proxy.isSupplier(manufacturer));
        assertTrue(proxy.hasRole(proxy.MANUFACTURER_ROLE(), manufacturer));
    }

    function testSupplierOnboardingRejectAndDeactivate() public {
        // Certifier requests onboarding
        vm.prank(certifier);
        proxy.requestSupplierOnboarding("CERTIFIER_ROLE", "CertifyCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash5678");
        // Super admin rejects
        vm.prank(superAdmin);
        proxy.rejectSupplierOnboarding(certifier);
        (, , , , , , AluminiumPassport.OnboardingStatus status,,) = proxy.onboardingRequests(certifier);
        assertEq(uint256(status), uint256(AluminiumPassport.OnboardingStatus.Rejected));
        // Super admin can deactivate (should revert if not a supplier)
        vm.prank(superAdmin);
        vm.expectRevert("Not a supplier");
        proxy.deactivateSupplier(certifier);
    }

    function testDeactivateSupplier() public {
        // Onboard and approve
        vm.prank(recycler);
        proxy.requestSupplierOnboarding("RECYCLER_ROLE", "RecycleCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash9999");
        vm.prank(superAdmin);
        proxy.approveSupplierOnboarding(recycler);
        // Deactivate
        vm.prank(superAdmin);
        proxy.deactivateSupplier(recycler);
        assertFalse(proxy.isSupplier(recycler));
        (, , , , , , AluminiumPassport.OnboardingStatus status,,) = proxy.onboardingRequests(recycler);
        assertEq(uint256(status), uint256(AluminiumPassport.OnboardingStatus.Deactivated));
    }

    function testCreateAndUpdatePassport() public {
        // Onboard manufacturer
        vm.prank(manufacturer);
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
        vm.prank(superAdmin);
        proxy.approveSupplierOnboarding(manufacturer);
        // Manufacturer creates passport
        vm.prank(manufacturer);
        proxy.createPassport(
            "PASS001",
            "Australia",
            "AluCo",
            "Al-99.7%",
            "CertifyCo",
            "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000",
            88,
            30
        );
        // Update passport
        vm.prank(manufacturer);
        proxy.updatePassport("PASS001", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1111", 90, 35);
        // Deactivate passport
        vm.prank(admin);
        proxy.deactivatePassport("PASS001");
        (, , , , , , , bool isActive, , , , ) = proxy.getPassport("PASS001");
        assertFalse(isActive);
    }

    function testRoleRevocation() public {
        // Onboard and approve
        vm.prank(manufacturer);
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
        vm.prank(superAdmin);
        proxy.approveSupplierOnboarding(manufacturer);
        // Super admin revokes role
        vm.prank(superAdmin);
        proxy.revokeRoleFrom(manufacturer, proxy.MANUFACTURER_ROLE());
        assertFalse(proxy.hasRole(proxy.MANUFACTURER_ROLE(), manufacturer));
    }

    function testInputValidation() public {
        // Empty company name
        vm.prank(manufacturer);
        vm.expectRevert("Company name required");
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
        // Invalid IPFS hash
        vm.prank(manufacturer);
        vm.expectRevert("Invalid IPFS hash");
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AluCo", "badHash");
        // Empty passportId
        vm.prank(manufacturer);
        vm.expectRevert("passportId required");
        proxy.createPassport("", "Australia", "AluCo", "Al-99.7%", "CertifyCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000", 88, 30);
        // Too many certifications
        vm.prank(manufacturer);
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
        vm.prank(superAdmin);
        proxy.approveSupplierOnboarding(manufacturer);
        vm.prank(manufacturer);
        proxy.createPassport("PASS002", "Australia", "AluCo", "Al-99.7%", "CertifyCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000", 88, 30);
        for (uint i = 0; i < 50; i++) {
            proxy.addCertification("PASS002", string(abi.encodePacked("Cert", i)));
        }
        vm.expectRevert("Too many certifications");
        proxy.addCertification("PASS002", "OverflowCert");
    }
}