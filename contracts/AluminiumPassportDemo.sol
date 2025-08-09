// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/AccessControl.sol";

contract AluminiumPassportDemo is AccessControl {
    // Roles
    bytes32 public constant SUPER_ADMIN_ROLE = keccak256("SUPER_ADMIN_ROLE");
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant REFINER_ROLE = keccak256("REFINER_ROLE");
    bytes32 public constant IMPORTER_ROLE = keccak256("IMPORTER_ROLE");
    bytes32 public constant AUDITOR_ROLE = keccak256("AUDITOR_ROLE");
    bytes32 public constant RECYCLER_ROLE = keccak256("RECYCLER_ROLE");
    bytes32 public constant MINER_ROLE = keccak256("MINER_ROLE");
    bytes32 public constant ALLOY_PRODUCER_ROLE = keccak256("ALLOY_PRODUCER_ROLE");
    bytes32 public constant MANUFACTURER_ROLE = keccak256("MANUFACTURER_ROLE");
    bytes32 public constant DISTRIBUTOR_ROLE = keccak256("DISTRIBUTOR_ROLE");
    bytes32 public constant SERVICE_PROVIDER_ROLE = keccak256("SERVICE_PROVIDER_ROLE");
    bytes32 public constant REGULATOR_ROLE = keccak256("REGULATOR_ROLE");

    enum OnboardingStatus { None, Pending, Approved, Rejected }

    // Onboarding
    struct OnboardingRequest {
        string orgId;
        address wallet;
        string kycCid;
        string metaCid;
        bytes32[] rolesRequested;
        bool exists;
        bool approved;
        uint256 requestedAt;
        uint256 approvedAt;
    }

    // Upstream batches (MINER)
    struct UpstreamBatch {
        string batchId;   // unique id from miner
        string cid;       // IPFS CID with origin & ESG data
        address registeredBy;
        uint256 timestamp;
        bool exists;
    }

    // Passport and related data
    struct Passport {
        uint256 id;
        string orgId;
        string upstreamBatchId; // links to UpstreamBatch
        string metaCid; // base metadata (e.g., CoA, CFP JSON CIDs bundled)
        uint256 parentId; // 0 if primary; non-zero for secondary spawned passports
        address createdBy;
        uint256 createdAt;
        bool exists;
    }

    struct StageEntry {
        string stage; // e.g., "smelting", "casting", "billet", etc.
        string cid;   // IPFS CID of stage metadata
        address addedBy;
        uint256 timestamp;
    }

    struct PlacedOnMarketRecord {
        string countryCode;
        string dateISO; // ISO 8601 date string
        string cid;     // IPFS evidence
        address recordedBy;
        uint256 timestamp;
        bool exists;
    }

    struct Attestation {
        string cid; // IPFS CID for attestation evidence
        address attestedBy;
        uint256 timestamp;
    }

    struct RecoveryRecord {
        uint8 recoveryPercent; // 0-100
        string quality;        // e.g., grade label
        string cid;            // IPFS evidence
        address recordedBy;
        uint256 timestamp;
    }

    // Storage
    mapping(string => OnboardingRequest) public onboardingByOrg; // orgId => request
    mapping(address => string) public walletOrg;                 // wallet => orgId
    mapping(string => bool) public orgSuspended;                 // orgId => suspended

    mapping(string => UpstreamBatch) public upstreamBatches;     // batchId => batch
    mapping(uint256 => Passport) public passports;               // passportId => Passport
    mapping(uint256 => StageEntry[]) public stagesByPassport;    // stage history
    mapping(uint256 => PlacedOnMarketRecord) public placedOnMarket; // Importer record
    mapping(uint256 => Attestation[]) public attestationsByPassport; // Auditor attestations
    mapping(uint256 => RecoveryRecord[]) public recoveryByPassport;  // Recycler recovery logs

    uint256 public nextPassportId;

    // Events
    event OnboardingRequested(string indexed orgId, address indexed wallet, bytes32[] roles, string kycCid, string metaCid, uint256 timestamp);
    event OnboardingApproved(string indexed orgId, address indexed wallet, bytes32[] roles, address indexed approvedBy, uint256 timestamp);
    event OnboardingRejected(string indexed orgId, address indexed wallet, address indexed rejectedBy, uint256 timestamp);
    event OrgSuspended(string indexed orgId, address indexed by, string reasonCid, uint256 timestamp);
    event OrgUnsuspended(string indexed orgId, address indexed by, uint256 timestamp);

    event UpstreamBatchRegistered(string indexed batchId, string cid, address indexed registeredBy, uint256 timestamp);
    event PassportCreated(uint256 indexed passportId, string indexed orgId, string metaCid, address indexed createdBy, uint256 timestamp);
    event UpstreamLinked(uint256 indexed passportId, string indexed upstreamBatchId);
    event StageDataAppended(uint256 indexed passportId, string stage, string cid, address indexed addedBy, uint256 timestamp);
    event PlacedOnMarket(uint256 indexed passportId, string countryCode, string dateISO, string cid, address indexed recordedBy, uint256 timestamp);
    event AttestationAdded(uint256 indexed passportId, string cid, address indexed attestedBy, uint256 timestamp);
    event RecoveryRecorded(uint256 indexed passportId, uint8 recoveryPercent, string quality, string cid, address indexed recordedBy, uint256 timestamp);
    event SecondaryPassportSpawned(uint256 indexed parentId, uint256 indexed newPassportId, string metaCid, address indexed createdBy, uint256 timestamp);

    constructor(address superAdmin, address admin) {
        _grantRole(DEFAULT_ADMIN_ROLE, superAdmin);
        _grantRole(SUPER_ADMIN_ROLE, superAdmin);
        _grantRole(ADMIN_ROLE, admin);
        nextPassportId = 1;
    }

    // --- Onboarding ---
    function requestOnboarding(
        string memory orgId,
        address wallet,
        string memory kycCid,
        string memory metaCid,
        bytes32[] memory rolesRequested
    ) external onlyRole(ADMIN_ROLE) {
        require(wallet != address(0), "wallet required");
        require(bytes(orgId).length > 0, "orgId required");
        OnboardingRequest storage r = onboardingByOrg[orgId];
        r.orgId = orgId;
        r.wallet = wallet;
        r.kycCid = kycCid;
        r.metaCid = metaCid;
        r.exists = true;
        r.approved = false;
        r.rolesRequested = rolesRequested;
        r.requestedAt = block.timestamp;
        r.approvedAt = 0;
        emit OnboardingRequested(orgId, wallet, rolesRequested, kycCid, metaCid, block.timestamp);
    }

    function rejectOnboarding(string memory orgId) external onlyRole(SUPER_ADMIN_ROLE) {
        OnboardingRequest storage r = onboardingByOrg[orgId];
        require(r.exists, "no request");
        require(!r.approved, "already approved");
        address wallet = r.wallet;
        r.exists = false;
        emit OnboardingRejected(orgId, wallet, msg.sender, block.timestamp);
    }

    function approveOnboarding(
        string memory orgId,
        bytes32[] memory rolesToGrant
    ) external onlyRole(SUPER_ADMIN_ROLE) {
        OnboardingRequest storage r = onboardingByOrg[orgId];
        require(r.exists, "no request");
        require(!r.approved, "already approved");
        r.approved = true;
        r.approvedAt = block.timestamp;
        walletOrg[r.wallet] = orgId;
        for (uint256 i = 0; i < rolesToGrant.length; i++) {
            _grantRole(rolesToGrant[i], r.wallet);
        }
        emit OnboardingApproved(orgId, r.wallet, rolesToGrant, msg.sender, block.timestamp);
    }

    function suspendOrg(string memory orgId, string memory reasonCid) external {
        require(hasRole(SUPER_ADMIN_ROLE, msg.sender) || hasRole(REGULATOR_ROLE, msg.sender), "Not allowed");
        require(bytes(orgId).length > 0, "orgId required");
        orgSuspended[orgId] = true;
        emit OrgSuspended(orgId, msg.sender, reasonCid, block.timestamp);
    }

    function unsuspendOrg(string memory orgId) external {
        require(hasRole(SUPER_ADMIN_ROLE, msg.sender) || hasRole(REGULATOR_ROLE, msg.sender), "Not allowed");
        require(bytes(orgId).length > 0, "orgId required");
        orgSuspended[orgId] = false;
        emit OrgUnsuspended(orgId, msg.sender, block.timestamp);
    }

    function _requireNotSuspended(address actor) internal view {
        string memory orgId = walletOrg[actor];
        if (bytes(orgId).length > 0) {
            require(!orgSuspended[orgId], "org suspended");
        }
    }

    // --- Upstream (Miner) ---
    function registerUpstreamBatch(string memory batchId, string memory cid) external onlyRole(MINER_ROLE) {
        require(bytes(batchId).length > 0, "batchId required");
        require(!upstreamBatches[batchId].exists, "batch exists");
        upstreamBatches[batchId] = UpstreamBatch({
            batchId: batchId,
            cid: cid,
            registeredBy: msg.sender,
            timestamp: block.timestamp,
            exists: true
        });
        emit UpstreamBatchRegistered(batchId, cid, msg.sender, block.timestamp);
    }

    // --- Passport ---
    // createPassport(orgId, upstreamBatchId, metaCid) (REFINER or IMPORTER)
    function createPassport(
        string memory orgId,
        string memory upstreamBatchId,
        string memory metaCid
    ) external returns (uint256 passportId) {
        _requireNotSuspended(msg.sender);
        require(
            hasRole(REFINER_ROLE, msg.sender) || hasRole(IMPORTER_ROLE, msg.sender),
            "Not refiner/importer"
        );
        require(bytes(orgId).length > 0, "orgId required");
        if (bytes(upstreamBatchId).length > 0) {
            require(upstreamBatches[upstreamBatchId].exists, "no upstream");
        }
        passportId = nextPassportId++;
        passports[passportId] = Passport({
            id: passportId,
            orgId: orgId,
            upstreamBatchId: upstreamBatchId,
            metaCid: metaCid,
            parentId: 0,
            createdBy: msg.sender,
            createdAt: block.timestamp,
            exists: true
        });
        emit PassportCreated(passportId, orgId, metaCid, msg.sender, block.timestamp);
        if (bytes(upstreamBatchId).length > 0) {
            emit UpstreamLinked(passportId, upstreamBatchId);
        }
    }

    // appendStageData(passportId, stage, cid) (role-gated)
    function appendStageData(uint256 passportId, string memory stage, string memory cid) external {
        _requireNotSuspended(msg.sender);
        require(passports[passportId].exists, "no passport");
        require(
            hasRole(MINER_ROLE, msg.sender) ||
            hasRole(REFINER_ROLE, msg.sender) ||
            hasRole(ALLOY_PRODUCER_ROLE, msg.sender) ||
            hasRole(MANUFACTURER_ROLE, msg.sender) ||
            hasRole(IMPORTER_ROLE, msg.sender) ||
            hasRole(DISTRIBUTOR_ROLE, msg.sender) ||
            hasRole(SERVICE_PROVIDER_ROLE, msg.sender) ||
            hasRole(RECYCLER_ROLE, msg.sender) ||
            hasRole(AUDITOR_ROLE, msg.sender),
            "Not allowed"
        );
        stagesByPassport[passportId].push(StageEntry({
            stage: stage,
            cid: cid,
            addedBy: msg.sender,
            timestamp: block.timestamp
        }));
        emit StageDataAppended(passportId, stage, cid, msg.sender, block.timestamp);
    }

    // recordPlacedOnMarket(passportId, countryCode, dateISO, cid) (IMPORTER)
    function recordPlacedOnMarket(
        uint256 passportId,
        string memory countryCode,
        string memory dateISO,
        string memory cid
    ) external {
        _requireNotSuspended(msg.sender);
        require(passports[passportId].exists, "no passport");
        require(hasRole(IMPORTER_ROLE, msg.sender), "Not importer");
        placedOnMarket[passportId] = PlacedOnMarketRecord({
            countryCode: countryCode,
            dateISO: dateISO,
            cid: cid,
            recordedBy: msg.sender,
            timestamp: block.timestamp,
            exists: true
        });
        emit PlacedOnMarket(passportId, countryCode, dateISO, cid, msg.sender, block.timestamp);
    }

    // addAttestation(passportId, cid) (AUDITOR)
    function addAttestation(uint256 passportId, string memory cid) external {
        _requireNotSuspended(msg.sender);
        require(passports[passportId].exists, "no passport");
        require(hasRole(AUDITOR_ROLE, msg.sender), "Not auditor");
        attestationsByPassport[passportId].push(Attestation({
            cid: cid,
            attestedBy: msg.sender,
            timestamp: block.timestamp
        }));
        emit AttestationAdded(passportId, cid, msg.sender, block.timestamp);
    }

    // Recycler recovery logging
    function recordRecovery(uint256 passportId, uint8 recoveryPercent, string memory quality, string memory cid) external {
        _requireNotSuspended(msg.sender);
        require(passports[passportId].exists, "no passport");
        require(hasRole(RECYCLER_ROLE, msg.sender), "Not recycler");
        require(recoveryPercent <= 100, "percent > 100");
        recoveryByPassport[passportId].push(RecoveryRecord({
            recoveryPercent: recoveryPercent,
            quality: quality,
            cid: cid,
            recordedBy: msg.sender,
            timestamp: block.timestamp
        }));
        emit RecoveryRecorded(passportId, recoveryPercent, quality, cid, msg.sender, block.timestamp);
    }

    // spawnSecondaryPassport(parentId, metaCid) (RECYCLER|REFINER)
    function spawnSecondaryPassport(uint256 parentId, string memory metaCid) external returns (uint256 newPassportId) {
        _requireNotSuspended(msg.sender);
        require(passports[parentId].exists, "no parent");
        require(
            hasRole(RECYCLER_ROLE, msg.sender) || hasRole(REFINER_ROLE, msg.sender),
            "Not recycler/refiner"
        );
        newPassportId = nextPassportId++;
        passports[newPassportId] = Passport({
            id: newPassportId,
            orgId: passports[parentId].orgId,
            upstreamBatchId: passports[parentId].upstreamBatchId,
            metaCid: metaCid,
            parentId: parentId,
            createdBy: msg.sender,
            createdAt: block.timestamp,
            exists: true
        });
        emit SecondaryPassportSpawned(parentId, newPassportId, metaCid, msg.sender, block.timestamp);
    }

    // Public read model for QR scan route
    function getPublicView(uint256 passportId) external view returns (
        string memory orgId,
        string memory upstreamBatchId,
        string memory passportMetaCid,
        bool placed,
        string memory countryCode,
        string memory dateISO,
        string memory placedCid,
        bool hasAttestation
    ) {
        Passport storage p = passports[passportId];
        require(p.exists, "no passport");
        orgId = p.orgId;
        upstreamBatchId = p.upstreamBatchId;
        passportMetaCid = p.metaCid;
        PlacedOnMarketRecord storage m = placedOnMarket[passportId];
        placed = m.exists;
        countryCode = m.countryCode;
        dateISO = m.dateISO;
        placedCid = m.cid;
        hasAttestation = attestationsByPassport[passportId].length > 0;
    }
}