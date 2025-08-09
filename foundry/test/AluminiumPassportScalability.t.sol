// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../../contracts/AluminiumPassport.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import "@openzeppelin/contracts/access/IAccessControl.sol";

contract AluminiumPassportScalabilityTest is Test {
    AluminiumPassport logic;
    AluminiumPassport proxy;
    address superAdmin = address(0xA11CE);
    address admin = address(0xBEEF);
    
    // Test addresses for batch operations
    address[] suppliers;
    bytes32[] roles;

    function setUp() public {
        logic = new AluminiumPassport();
        bytes memory data = abi.encodeWithSignature("initialize(address,address)", superAdmin, admin);
        ERC1967Proxy proxyContract = new ERC1967Proxy(address(logic), data);
        proxy = AluminiumPassport(address(proxyContract));
        
        // Setup test suppliers
        for (uint256 i = 0; i < 50; i++) {
            suppliers.push(address(uint160(0x1000 + i)));
            roles.push(proxy.PRODUCT_MANUFACTURER_ROLE());
        }
    }

    function testBatchApproveSuppliers() public {
        // Create pending requests first
        for (uint256 i = 0; i < 10; i++) {
            vm.prank(suppliers[i]);
            proxy.requestSupplierOnboarding(
                "PRODUCT_MANUFACTURER_ROLE",
                string(abi.encodePacked("Supplier", vm.toString(i))),
                "QmTestHash"
            );
        }
        
        // Batch approve
        address[] memory suppliersToApprove = new address[](10);
        bytes32[] memory rolesToGrant = new bytes32[](10);
        
        for (uint256 i = 0; i < 10; i++) {
            suppliersToApprove[i] = suppliers[i];
            rolesToGrant[i] = proxy.PRODUCT_MANUFACTURER_ROLE();
        }
        
        vm.prank(superAdmin);
        proxy.batchApproveSuppliers(suppliersToApprove, rolesToGrant);
        
        // Verify all are approved
        for (uint256 i = 0; i < 10; i++) {
            assertTrue(proxy.isSupplier(suppliers[i]));
            assertTrue(proxy.hasRole(proxy.PRODUCT_MANUFACTURER_ROLE(), suppliers[i]));
        }
    }

    function testBatchRejectSuppliers() public {
        // Create pending requests first
        for (uint256 i = 0; i < 10; i++) {
            vm.prank(suppliers[i]);
            proxy.requestSupplierOnboarding(
                "PRODUCT_MANUFACTURER_ROLE",
                string(abi.encodePacked("Supplier", vm.toString(i))),
                "QmTestHash"
            );
        }
        
        // Batch reject
        address[] memory suppliersToReject = new address[](10);
        for (uint256 i = 0; i < 10; i++) {
            suppliersToReject[i] = suppliers[i];
        }
        
        vm.prank(superAdmin);
        proxy.batchRejectSuppliers(suppliersToReject);
        
        // Verify all are rejected
        for (uint256 i = 0; i < 10; i++) {
            (,,,,,, AluminiumPassport.OnboardingStatus status,,) = proxy.onboardingRequests(suppliers[i]);
            assertEq(uint8(status), 2); // Rejected
        }
    }

    function testPreApprovedSupplierSystem() public {
        // Add supplier to pre-approved list
        vm.prank(superAdmin);
        proxy.addPreApprovedSupplier(suppliers[0], proxy.PRODUCT_MANUFACTURER_ROLE());
        
        // Auto-approve onboarding
        vm.prank(suppliers[0]);
        proxy.autoApproveSupplierOnboarding(
            "PRODUCT_MANUFACTURER_ROLE",
            "PreApprovedSupplier",
            "QmTestHash",
            "company.com"
        );
        
        // Verify auto-approval
        assertTrue(proxy.isSupplier(suppliers[0]));
        assertTrue(proxy.hasRole(proxy.PRODUCT_MANUFACTURER_ROLE(), suppliers[0]));
        
        (,,,,,, AluminiumPassport.OnboardingStatus status,,) = proxy.onboardingRequests(suppliers[0]);
        assertEq(uint8(status), 1); // Approved
    }

    function testBulkPreApprovedSuppliers() public {
        // Setup arrays for bulk operation
        address[] memory suppliersToPreApprove = new address[](20);
        bytes32[] memory rolesToPreApprove = new bytes32[](20);
        
        for (uint256 i = 0; i < 20; i++) {
            suppliersToPreApprove[i] = suppliers[i];
            rolesToPreApprove[i] = proxy.PRODUCT_MANUFACTURER_ROLE();
        }
        
        // Bulk add pre-approved suppliers
        vm.prank(superAdmin);
        proxy.bulkAddPreApprovedSuppliers(suppliersToPreApprove, rolesToPreApprove);
        
        // Verify all are pre-approved
        for (uint256 i = 0; i < 20; i++) {
            assertTrue(proxy.preApprovedSuppliers(suppliers[i]));
        }
    }

    function testBulkDeactivateSuppliers() public {
        // First approve some suppliers
        for (uint256 i = 0; i < 10; i++) {
            vm.prank(superAdmin);
            proxy.addPreApprovedSupplier(suppliers[i], proxy.PRODUCT_MANUFACTURER_ROLE());
            
            vm.prank(suppliers[i]);
            proxy.autoApproveSupplierOnboarding(
                "PRODUCT_MANUFACTURER_ROLE",
                string(abi.encodePacked("Supplier", vm.toString(i))),
                "QmTestHash",
                "company.com"
            );
        }
        
        // Bulk deactivate
        address[] memory suppliersToDeactivate = new address[](10);
        for (uint256 i = 0; i < 10; i++) {
            suppliersToDeactivate[i] = suppliers[i];
        }
        
        vm.prank(superAdmin);
        proxy.bulkDeactivateSuppliers(suppliersToDeactivate);
        
        // Verify all are deactivated
        for (uint256 i = 0; i < 10; i++) {
            assertFalse(proxy.isSupplier(suppliers[i]));
        }
    }

    function testSupplierStatusCheck() public {
        // Add and approve a supplier
        vm.prank(superAdmin);
        proxy.addPreApprovedSupplier(suppliers[0], proxy.PRODUCT_MANUFACTURER_ROLE());
        
        vm.prank(suppliers[0]);
        proxy.autoApproveSupplierOnboarding(
            "PRODUCT_MANUFACTURER_ROLE",
            "TestSupplier",
            "QmTestHash",
            "company.com"
        );
        
        // Check status
        (bool isActive, bool hasRole) = proxy.getSupplierStatus(suppliers[0], proxy.PRODUCT_MANUFACTURER_ROLE());
        assertTrue(isActive);
        assertTrue(hasRole);
    }

    function testGasOptimizationForLargeScale() public {
        // Test gas usage for batch operations
        uint256 gasBefore = gasleft();
        
        // Batch approve 50 suppliers
        address[] memory suppliersToApprove = new address[](50);
        bytes32[] memory rolesToGrant = new bytes32[](50);
        
        for (uint256 i = 0; i < 50; i++) {
            suppliersToApprove[i] = suppliers[i];
            rolesToGrant[i] = proxy.PRODUCT_MANUFACTURER_ROLE();
        }
        
        vm.prank(superAdmin);
        proxy.batchApproveSuppliers(suppliersToApprove, rolesToGrant);
        
        uint256 gasUsed = gasBefore - gasleft();
        // console.log("Gas used for batch approving 50 suppliers:", gasUsed);
        
        // This should be much more efficient than 50 individual transactions
        assertTrue(gasUsed < 5000000); // Should be under 5M gas
    }

    function testBatchSizeLimits() public {
        // Test that batch size limits are enforced
        address[] memory tooManySuppliers = new address[](51);
        bytes32[] memory tooManyRoles = new bytes32[](51);
        
        for (uint256 i = 0; i < 51; i++) {
            tooManySuppliers[i] = suppliers[i % 50];
            tooManyRoles[i] = proxy.PRODUCT_MANUFACTURER_ROLE();
        }
        
        vm.prank(superAdmin);
        vm.expectRevert("Max 50 suppliers per batch");
        proxy.batchApproveSuppliers(tooManySuppliers, tooManyRoles);
    }

    function testPreApprovedDomainSystem() public {
        // Add domain to pre-approved list
        vm.prank(superAdmin);
        proxy.addPreApprovedDomain("trusted-company.com");
        
        // Test auto-approval with domain
        vm.prank(suppliers[0]);
        proxy.autoApproveSupplierOnboarding(
            "PRODUCT_MANUFACTURER_ROLE",
            "TrustedCompany",
            "QmTestHash",
            "trusted-company.com"
        );
        
        // Verify approval
        assertTrue(proxy.isSupplier(suppliers[0]));
    }

    function testScalabilitySimulation() public {
        // Simulate onboarding 3000 suppliers in batches
        uint256 totalSuppliers = 3000;
        uint256 batchSize = 50;
        uint256 totalBatches = totalSuppliers / batchSize;
        
        // console.log("Simulating onboarding of", totalSuppliers, "suppliers in", totalBatches, "batches");
        
        for (uint256 batch = 0; batch < totalBatches; batch++) {
            address[] memory batchSuppliers = new address[](batchSize);
            bytes32[] memory batchRoles = new bytes32[](batchSize);
            
            for (uint256 i = 0; i < batchSize; i++) {
                uint256 supplierIndex = batch * batchSize + i;
                batchSuppliers[i] = address(uint160(0x2000 + supplierIndex));
                batchRoles[i] = proxy.PRODUCT_MANUFACTURER_ROLE();
            }
            
            // Simulate batch approval
            vm.prank(superAdmin);
            proxy.bulkAddPreApprovedSuppliers(batchSuppliers, batchRoles);
            
            if (batch % 10 == 0) {
                // console.log("Completed batch", batch, "of", totalBatches);
            }
        }
        
        // console.log("Successfully simulated onboarding of", totalSuppliers, "suppliers");
    }
} 