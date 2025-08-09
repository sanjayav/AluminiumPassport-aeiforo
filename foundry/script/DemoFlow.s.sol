// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import "../../contracts/AluminiumPassportDemo.sol";

contract DemoFlow is Script {
    // Seed wallets (replace with real addresses for demo)
    address constant SUPER_ADMIN = 0x0000000000000000000000000000000000000a11;
    address constant ADMIN       = 0x0000000000000000000000000000000000000AD0;
    address constant MINER       = 0x0000000000000000000000000000000000000Mi9;
    address constant REFINER     = 0x0000000000000000000000000000000000000Rf1;
    address constant IMPORTER    = 0x0000000000000000000000000000000000000Im0;
    address constant AUDITOR     = 0x0000000000000000000000000000000000000Au7;
    address constant RECYCLER    = 0x0000000000000000000000000000000000000Rc5;

    function run() external {
        uint256 deployerKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerKey);

        // Deploy demo contract with super admin and admin
        AluminiumPassportDemo demo = new AluminiumPassportDemo(SUPER_ADMIN, ADMIN);

        // SuperAdmin grants base roles to seed wallets (importer via onboarding)
        vm.prank(SUPER_ADMIN);
        demo.grantRole(demo.MINER_ROLE(), MINER);
        vm.prank(SUPER_ADMIN);
        demo.grantRole(demo.REFINER_ROLE(), REFINER);
        vm.prank(SUPER_ADMIN);
        demo.grantRole(demo.AUDITOR_ROLE(), AUDITOR);
        vm.prank(SUPER_ADMIN);
        demo.grantRole(demo.RECYCLER_ROLE(), RECYCLER);

        // Admin submits onboarding for Importer
        bytes32[] memory importerRoles = new bytes32[](1);
        importerRoles[0] = demo.IMPORTER_ROLE();
        vm.prank(ADMIN);
        demo.requestOnboarding("EGA-IMPORT", IMPORTER, "ipfs://kycCID", "ipfs://metaCID", importerRoles);

        // SuperAdmin approves onboarding, granting IMPORTER role
        vm.prank(SUPER_ADMIN);
        demo.approveOnboarding("EGA-IMPORT", importerRoles);

        // Miner registers upstream batch (origin & ESG data)
        vm.prank(MINER);
        demo.registerUpstreamBatch("BATCH-ALU-001", "ipfs://upstream_esg_cid");

        // Refiner creates a passport linking upstream batch with CoA/CFP metadata CID
        vm.prank(REFINER);
        uint256 passportId = demo.createPassport("EGA-REFINE", "BATCH-ALU-001", "ipfs://coa_cfp_cid");

        // Importer records placed on EU market
        vm.prank(IMPORTER);
        demo.recordPlacedOnMarket(passportId, "DE", "2025-09-01", "ipfs://shipment_cid");

        // Auditor adds attestation
        vm.prank(AUDITOR);
        demo.addAttestation(passportId, "ipfs://attestation_cid");

        // Recycler logs recovery and spawns secondary passport
        vm.prank(RECYCLER);
        demo.recordRecovery(passportId, 82, "A-Quality", "ipfs://recovery_cid");
        vm.prank(RECYCLER);
        demo.spawnSecondaryPassport(passportId, "ipfs://secondary_meta_cid");

        // Public read for QR view
        (
            string memory orgId,
            string memory upstreamBatchId,
            string memory passportMetaCid,
            bool placed,
            string memory countryCode,
            string memory dateISO,
            string memory placedCid,
            bool hasAttestation
        ) = demo.getPublicView(passportId);
        // Values can be printed with console if desired.

        vm.stopBroadcast();
    }
}