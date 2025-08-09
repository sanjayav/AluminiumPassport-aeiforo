// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Test.sol";
import "../../contracts/AluminiumPassportDemo.sol";

contract AluminiumPassportDemoTest is Test {
    AluminiumPassportDemo private demo;

    address private superAdmin = makeAddr("superAdmin");
    address private admin = makeAddr("admin");
    address private miner = makeAddr("miner");
    address private refiner = makeAddr("refiner");
    address private importer = makeAddr("importer");
    address private auditor = makeAddr("auditor");
    address private recycler = makeAddr("recycler");

    function setUp() public {
        demo = new AluminiumPassportDemo(superAdmin, admin);

        // Grant base roles (importer will be onboarded inside the test)
        vm.startPrank(superAdmin);
        demo.grantRole(demo.MINER_ROLE(), miner);
        demo.grantRole(demo.REFINER_ROLE(), refiner);
        demo.grantRole(demo.AUDITOR_ROLE(), auditor);
        demo.grantRole(demo.RECYCLER_ROLE(), recycler);
        vm.stopPrank();
    }

    function testDemoLifecycle() public {
        // Onboard importer
        bytes32[] memory importerRoles = new bytes32[](1);
        importerRoles[0] = demo.IMPORTER_ROLE();
        vm.prank(admin);
        demo.requestOnboarding("ORG-IMPORT", importer, "ipfs://kycCID", "ipfs://metaCID", importerRoles);
        vm.prank(superAdmin);
        demo.approveOnboarding("ORG-IMPORT", importerRoles);

        // Miner registers upstream batch
        vm.prank(miner);
        demo.registerUpstreamBatch("BATCH-001", "ipfs://upstream_esg_cid");

        // Refiner creates passport linking upstream batch
        vm.prank(refiner);
        uint256 passportId = demo.createPassport("ORG-REF", "BATCH-001", "ipfs://coa_cfp_cid");

        // Importer records placed on market
        vm.prank(importer);
        demo.recordPlacedOnMarket(passportId, "DE", "2025-09-01", "ipfs://shipment_cid");

        // Auditor adds attestation
        vm.prank(auditor);
        demo.addAttestation(passportId, "ipfs://attestation_cid");

        // Recycler records recovery and spawns secondary
        vm.prank(recycler);
        demo.recordRecovery(passportId, 80, "A-Quality", "ipfs://recovery_cid");
        vm.prank(recycler);
        uint256 secondaryId = demo.spawnSecondaryPassport(passportId, "ipfs://secondary_meta_cid");

        // Assertions
        // Upstream batch exists
        (string memory bId, string memory bCid, address bBy, uint256 bTs, bool bExists) = demo.upstreamBatches("BATCH-001");
        assertEq(bId, "BATCH-001");
        assertEq(bCid, "ipfs://upstream_esg_cid");
        assertEq(bBy, miner);
        assertTrue(bExists);
        assertGt(bTs, 0);

        // Primary passport stored and linked
        (uint256 id, string memory orgId, string memory upId, string memory metaCid, uint256 parentId, address createdBy, uint256 createdAt, bool pExists) = demo.passports(passportId);
        assertEq(id, passportId);
        assertEq(orgId, "ORG-REF");
        assertEq(upId, "BATCH-001");
        assertEq(metaCid, "ipfs://coa_cfp_cid");
        assertEq(parentId, 0);
        assertEq(createdBy, refiner);
        assertTrue(pExists);
        assertGt(createdAt, 0);

        // Placed on market record
        (string memory country, string memory dateISO, string memory shipCid, address recordedBy, uint256 ts, bool pmExists) = demo.placedOnMarket(passportId);
        assertEq(country, "DE");
        assertEq(dateISO, "2025-09-01");
        assertEq(shipCid, "ipfs://shipment_cid");
        assertEq(recordedBy, importer);
        assertTrue(pmExists);
        assertGt(ts, 0);

        // Attestation present (index 0)
        (string memory attCid, address attBy, uint256 attTs) = demo.attestationsByPassport(passportId, 0);
        assertEq(attCid, "ipfs://attestation_cid");
        assertEq(attBy, auditor);
        assertGt(attTs, 0);

        // Recovery present (index 0)
        (uint8 recPct, string memory quality, string memory recCid, address recBy, uint256 recTs) = demo.recoveryByPassport(passportId, 0);
        assertEq(recPct, 80);
        assertEq(quality, "A-Quality");
        assertEq(recCid, "ipfs://recovery_cid");
        assertEq(recBy, recycler);
        assertGt(recTs, 0);

        // Secondary passport created
        (uint256 sId, string memory sOrgId, string memory sUpId, string memory sMetaCid, uint256 sParentId, address sBy, uint256 sAt, bool sExists) = demo.passports(secondaryId);
        assertEq(sId, secondaryId);
        assertEq(sOrgId, orgId); // inherits org
        assertEq(sUpId, upId);   // inherits upstream link
        assertEq(sMetaCid, "ipfs://secondary_meta_cid");
        assertEq(sParentId, passportId);
        assertEq(sBy, recycler);
        assertTrue(sExists);
        assertGt(sAt, 0);

        // Public view for primary
        (
            string memory povOrg,
            string memory povUp,
            string memory povMeta,
            bool placed,
            string memory povCountry,
            string memory povDate,
            string memory povPlacedCid,
            bool hasAtt
        ) = demo.getPublicView(passportId);
        assertEq(povOrg, "ORG-REF");
        assertEq(povUp, "BATCH-001");
        assertEq(povMeta, "ipfs://coa_cfp_cid");
        assertTrue(placed);
        assertEq(povCountry, "DE");
        assertEq(povDate, "2025-09-01");
        assertEq(povPlacedCid, "ipfs://shipment_cid");
        assertTrue(hasAtt);

        // Public view for secondary (not placed/attested)
        (, , string memory sPovMeta, bool sPlaced, , , , bool sHasAtt) = demo.getPublicView(secondaryId);
        assertEq(sPovMeta, "ipfs://secondary_meta_cid");
        assertFalse(sPlaced);
        assertFalse(sHasAtt);
    }
}
