// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../../contracts/AluminiumPassportMinimal.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import "@openzeppelin/contracts/access/IAccessControl.sol";

contract AluminiumPassportMinimalTest is Test {
    AluminiumPassportMinimal logic;
    AluminiumPassportMinimal proxy;
    address superAdmin = address(0xA11CE);
    address admin = address(0xBEEF);

    function setUp() public {
        logic = new AluminiumPassportMinimal();
        bytes memory data = abi.encodeWithSignature("initialize(address,address)", superAdmin, admin);
        ERC1967Proxy proxyContract = new ERC1967Proxy(address(logic), data);
        proxy = AluminiumPassportMinimal(address(proxyContract));
    }

    function testSuperAdminRole() public {
        assertTrue(proxy.hasRole(proxy.SUPER_ADMIN_ROLE(), superAdmin));
    }
    function testAdminRole() public {
        assertTrue(proxy.hasRole(proxy.ADMIN_ROLE(), admin));
    }
    function testVersion() public {
        assertEq(proxy.VERSION(), "minimal-roles-1.0.0");
    }



    function testCreateAndGetPassport() public {
        vm.prank(admin);
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
        (
            string memory pid,
            string memory origin,
            string memory manufacturer,
            string memory alloyComposition,
            string memory certifier,
            string memory ipfsHash,
            uint256 esgScore,
            uint256 recycledContent,
            bool isActive
        ) = proxy.getPassport("PASS001");
        assertEq(pid, "PASS001");
        assertEq(origin, "Australia");
        assertEq(manufacturer, "AluCo");
        assertEq(alloyComposition, "Al-99.7%");
        assertEq(certifier, "CertifyCo");
        assertEq(ipfsHash, "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash0000");
        assertEq(esgScore, 88);
        assertEq(recycledContent, 30);
        assertTrue(isActive);
    }

    function testSupplierOnboarding() public {
        vm.prank(superAdmin);
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
        (
            address supplier,
            string memory roleRequested,
            string memory companyName,
            string memory metadataIPFS,
            address requestedBy,
            , // requestedAt - unused
            AluminiumPassportMinimal.OnboardingStatus status,
            address approvedBy,
            uint256 approvedAt
        ) = proxy.onboardingRequests(superAdmin);
        assertEq(supplier, superAdmin);
        assertEq(roleRequested, "MANUFACTURER_ROLE");
        assertEq(companyName, "AluCo");
        assertEq(metadataIPFS, "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
        assertEq(requestedBy, superAdmin);
        assertEq(uint8(status), 0); // Pending
        assertEq(approvedBy, address(0));
        assertEq(approvedAt, 0);
    }

    function testPassportCreatedEvent() public {
        vm.prank(admin);
        vm.expectEmit(true, true, false, true);
        emit AluminiumPassportMinimal.PassportCreated(
            "PASS001",
            admin,
            "AluCo",
            "Australia",
            block.timestamp
        );
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
    }

    function testSupplierOnboardingRequestedEvent() public {
        vm.prank(superAdmin);
        vm.expectEmit(true, false, false, true);
        emit AluminiumPassportMinimal.SupplierOnboardingRequested(
            superAdmin,
            "MANUFACTURER_ROLE",
            "AluCo",
            superAdmin,
            block.timestamp
        );
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AluCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1234");
    }

    function testCreatePassportAccessControl() public {
        // Should succeed for admin
        vm.prank(admin);
        proxy.createPassport(
            "PASS002",
            "India",
            "OtherCo",
            "Al-99.5%",
            "CertifyCo",
            "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash1111",
            77,
            20
        );
        // Should succeed for super admin
        vm.prank(superAdmin);
        proxy.createPassport(
            "PASS003",
            "USA",
            "SuperCo",
            "Al-99.9%",
            "CertifyCo",
            "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash2222",
            99,
            50
        );
        // Should revert for others
        address notAdmin = address(0x1234);
        vm.prank(notAdmin);
        vm.expectRevert("Not authorized");
        proxy.createPassport(
            "FAIL",
            "Nowhere",
            "NoCo",
            "Al-0%",
            "NoCert",
            "QmFail",
            0,
            0
        );
    }

    function testRequestSupplierOnboardingAccessControl() public {
        // Should succeed for super admin
        vm.startPrank(superAdmin);
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "SuperCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash3333");
        vm.stopPrank();
        
        // Should revert for admin
        vm.startPrank(admin);
        vm.expectRevert(abi.encodeWithSelector(IAccessControl.AccessControlUnauthorizedAccount.selector, admin, proxy.SUPER_ADMIN_ROLE()));
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "AdminCo", "QmValidIPFSHashValidIPFSHashValidIPFSHashValidIPFSHash4444");
        vm.stopPrank();
        
        // Should revert for others
        address notAdmin = address(0x1234);
        vm.startPrank(notAdmin);
        vm.expectRevert(abi.encodeWithSelector(IAccessControl.AccessControlUnauthorizedAccount.selector, notAdmin, proxy.SUPER_ADMIN_ROLE()));
        proxy.requestSupplierOnboarding("MANUFACTURER_ROLE", "NoCo", "QmFail");
        vm.stopPrank();
    }
}