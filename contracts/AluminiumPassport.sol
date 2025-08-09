
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";

/// @title AluminiumPassport - Upgradeable, role-based, auditable aluminium passport contract
/// @notice Manages aluminium passports, supplier onboarding, and role-based access with upgradeability
contract AluminiumPassport is Initializable, UUPSUpgradeable, AccessControlUpgradeable, PausableUpgradeable {
    // --- Version ---
    string public constant VERSION = "2.1.0";

    // --- Roles ---
    bytes32 public constant SUPER_ADMIN_ROLE = keccak256("SUPER_ADMIN_ROLE");
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant MINER_ROLE = keccak256("MINER_ROLE");
    bytes32 public constant REFINER_ROLE = keccak256("REFINER_ROLE");
    bytes32 public constant ALLOY_PRODUCER_ROLE = keccak256("ALLOY_PRODUCER_ROLE");
    bytes32 public constant PRODUCT_MANUFACTURER_ROLE = keccak256("PRODUCT_MANUFACTURER_ROLE");
    bytes32 public constant DISTRIBUTOR_ROLE = keccak256("DISTRIBUTOR_ROLE");
    bytes32 public constant SERVICE_PROVIDER_ROLE = keccak256("SERVICE_PROVIDER_ROLE");
    bytes32 public constant RECYCLER_ROLE = keccak256("RECYCLER_ROLE");
    bytes32 public constant AUDITOR_ROLE = keccak256("AUDITOR_ROLE");
    bytes32 public constant REGULATOR_ROLE = keccak256("REGULATOR_ROLE");

    // --- Passport Struct ---
    struct Passport {
        string passportId;
        string origin;
        string manufacturer;
        string alloyComposition;
        string certifier;
        string ipfsHash; // Off-chain metadata
        uint256 esgScore;
        uint256 recycledContent;
        string[] certifications;
        string[] supplyChainSteps;
        address createdBy;
        uint256 createdAt;
        uint256 updatedAt;
        bool isActive;
    }

    // --- Supplier Onboarding Request ---
    enum OnboardingStatus { Pending, Approved, Rejected, Deactivated }
    struct SupplierOnboarding {
        address supplier;
        string roleRequested;
        string companyName;
        string metadataIPFS;
        address requestedBy;
        uint256 requestedAt;
        OnboardingStatus status;
        address approvedBy;
        uint256 approvedAt;
    }

    // --- Storage ---
    mapping(string => Passport) private passports;
    mapping(address => bool) public isSupplier;
    mapping(address => SupplierOnboarding) public onboardingRequests;
    string[] public allPassportIds;

    // --- Automated Approval System ---
    
    // Whitelist for pre-approved suppliers
    mapping(address => bool) public preApprovedSuppliers;
    mapping(string => bool) public preApprovedDomains; // For email domain-based approval
    mapping(bytes32 => bool) public preApprovedRoles; // For role-based auto-approval
    
    /// @notice Add supplier to pre-approved whitelist
    /// @param supplier Address to pre-approve
    /// @param role Role to automatically grant
    function addPreApprovedSupplier(
        address supplier,
        bytes32 role
    ) external onlyRole(SUPER_ADMIN_ROLE) {
        preApprovedSuppliers[supplier] = true;
        preApprovedRoles[role] = true;
    }
    
    /// @notice Remove supplier from pre-approved whitelist
    /// @param supplier Address to remove from pre-approval
    function removePreApprovedSupplier(address supplier) external onlyRole(SUPER_ADMIN_ROLE) {
        preApprovedSuppliers[supplier] = false;
    }
    
    /// @notice Add email domain to pre-approved list
    /// @param domain Email domain (e.g., "company.com")
    function addPreApprovedDomain(string calldata domain) external onlyRole(SUPER_ADMIN_ROLE) {
        preApprovedDomains[domain] = true;
    }
    
    /// @notice Automated supplier onboarding for pre-approved suppliers
    /// @param roleRequested Role being requested
    /// @param companyName Company name
    /// @param metadataIPFS IPFS hash of metadata
    /// @param emailDomain Email domain for verification
    function autoApproveSupplierOnboarding(
        string memory roleRequested,
        string memory companyName,
        string memory metadataIPFS,
        string memory emailDomain
    ) external whenNotPaused {
        require(preApprovedSuppliers[msg.sender] || preApprovedDomains[emailDomain], "Not pre-approved");
        
        bytes32 role = keccak256(abi.encodePacked(roleRequested));
        require(preApprovedRoles[role], "Role not pre-approved for auto-approval");
        
        // Auto-approve the supplier
        onboardingRequests[msg.sender] = SupplierOnboarding({
            supplier: msg.sender,
            roleRequested: roleRequested,
            companyName: companyName,
            metadataIPFS: metadataIPFS,
            requestedBy: msg.sender,
            requestedAt: block.timestamp,
            status: OnboardingStatus.Approved,
            approvedBy: address(0), // Auto-approved
            approvedAt: block.timestamp
        });
        
        // Grant the role
        _grantRole(role, msg.sender);
        isSupplier[msg.sender] = true;
        
        emit SupplierOnboardingRequested(msg.sender, roleRequested, companyName, msg.sender, block.timestamp);
        emit SupplierOnboardingApproved(msg.sender, address(0), block.timestamp);
    }
    
    /// @notice Bulk add pre-approved suppliers
    /// @param suppliers Array of supplier addresses
    /// @param roles Array of roles to pre-approve
    function bulkAddPreApprovedSuppliers(
        address[] calldata suppliers,
        bytes32[] calldata roles
    ) external onlyRole(SUPER_ADMIN_ROLE) {
        require(suppliers.length == roles.length, "Arrays length mismatch");
        require(suppliers.length <= 100, "Max 100 suppliers per batch");
        
        for (uint256 i = 0; i < suppliers.length; i++) {
            preApprovedSuppliers[suppliers[i]] = true;
            preApprovedRoles[roles[i]] = true;
        }
    }

    // --- Events ---
    event PassportCreated(string indexed passportId, address indexed createdBy, string manufacturer, string origin, uint256 timestamp);
    event PassportUpdated(string indexed passportId, address indexed updatedBy, uint256 timestamp);
    event PassportDeactivated(string indexed passportId, address indexed deactivatedBy, uint256 timestamp);
    event SupplierOnboardingRequested(address indexed supplier, string roleRequested, string companyName, address indexed requestedBy, uint256 timestamp);
    event SupplierOnboardingApproved(address indexed supplier, address indexed approvedBy, uint256 timestamp);
    event SupplierOnboardingRejected(address indexed supplier, address indexed rejectedBy, uint256 timestamp);
    event SupplierDeactivated(address indexed supplier, address indexed deactivatedBy, uint256 timestamp);
    event CertificationAdded(string indexed passportId, string certification, address indexed certifier, uint256 timestamp);
    event SupplyChainStepAdded(string indexed passportId, string step, address indexed addedBy, uint256 timestamp);
    event RoleRevoked(address indexed account, bytes32 indexed role, address indexed revokedBy, uint256 timestamp);
    event Paused(address indexed account, uint256 timestamp);
    event Unpaused(address indexed account, uint256 timestamp);
    event SuperAdminTransferred(address indexed oldSuperAdmin, address indexed newSuperAdmin, uint256 timestamp);

    // --- Modifiers ---
    modifier onlyRoleOrAdmin(bytes32 role) {
        require(
            hasRole(role, msg.sender) || hasRole(ADMIN_ROLE, msg.sender) || hasRole(SUPER_ADMIN_ROLE, msg.sender),
            "Not authorized"
        );
        _;
    }

    // --- Initializer ---
    /// @notice Initializes the contract and sets the super admin and admin
    /// @param superAdmin The address to be granted SUPER_ADMIN_ROLE
    /// @param admin The address to be granted ADMIN_ROLE
    function initialize(address superAdmin, address admin) public initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        __Pausable_init();
        _grantRole(DEFAULT_ADMIN_ROLE, superAdmin);
        _grantRole(SUPER_ADMIN_ROLE, superAdmin);
        _grantRole(ADMIN_ROLE, admin);
    }

    /// @notice Returns the contract version
    function getVersion() external pure returns (string memory) {
        return VERSION;
    }

    /// @notice Returns true if the contract is paused
    function isPaused() external view returns (bool) {
        return paused();
    }

    // --- UUPS Upgrade Authorization ---
    function _authorizeUpgrade(address newImplementation) internal override onlyRole(SUPER_ADMIN_ROLE) {}

    // --- Pausable ---
    /// @notice Pause the contract (SUPER_ADMIN or ADMIN)
    function pause() external onlyRoleOrAdmin(SUPER_ADMIN_ROLE) whenNotPaused {
        _pause();
        emit Paused(msg.sender, block.timestamp);
    }
    /// @notice Unpause the contract (SUPER_ADMIN or ADMIN)
    function unpause() external onlyRoleOrAdmin(SUPER_ADMIN_ROLE) whenPaused {
        _unpause();
        emit Unpaused(msg.sender, block.timestamp);
    }

    // --- Supplier Onboarding ---
    /// @notice Request onboarding as a supplier (manufacturer, recycler, certifier)
    /// @param roleRequested The role requested (must be one of the supply chain roles)
    /// @param companyName The name of the company
    /// @param metadataIPFS IPFS hash of supplier metadata
    function requestSupplierOnboarding(string memory roleRequested, string memory companyName, string memory metadataIPFS) external whenNotPaused {
        require(
            keccak256(bytes(roleRequested)) == keccak256("MINER_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("REFINER_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("ALLOY_PRODUCER_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("PRODUCT_MANUFACTURER_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("DISTRIBUTOR_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("SERVICE_PROVIDER_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("RECYCLER_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("AUDITOR_ROLE") ||
            keccak256(bytes(roleRequested)) == keccak256("REGULATOR_ROLE"),
            "Invalid supplier role"
        );
        require(!isSupplier[msg.sender], "Already a supplier");
        require(bytes(companyName).length > 0, "Company name required");
        require(bytes(metadataIPFS).length > 0, "IPFS hash required");
        require(_isValidIPFSHash(metadataIPFS), "Invalid IPFS hash");
        onboardingRequests[msg.sender] = SupplierOnboarding({
            supplier: msg.sender,
            roleRequested: roleRequested,
            companyName: companyName,
            metadataIPFS: metadataIPFS,
            requestedBy: msg.sender,
            requestedAt: block.timestamp,
            status: OnboardingStatus.Pending,
            approvedBy: address(0),
            approvedAt: 0
        });
        emit SupplierOnboardingRequested(msg.sender, roleRequested, companyName, msg.sender, block.timestamp);
    }

    /// @notice Approve a supplier onboarding request
    /// @param supplier The address of the supplier to approve
    function approveSupplierOnboarding(address supplier) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        SupplierOnboarding storage req = onboardingRequests[supplier];
        require(req.status == OnboardingStatus.Pending, "Not pending");
        req.status = OnboardingStatus.Approved;
        req.approvedBy = msg.sender;
        req.approvedAt = block.timestamp;
        isSupplier[supplier] = true;
        // Grant the requested role
        bytes32 role = _stringToRole(req.roleRequested);
        _grantRole(role, supplier);
        emit SupplierOnboardingApproved(supplier, msg.sender, block.timestamp);
    }

    /// @notice Reject a supplier onboarding request
    /// @param supplier The address of the supplier to reject
    function rejectSupplierOnboarding(address supplier) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        SupplierOnboarding storage req = onboardingRequests[supplier];
        require(req.status == OnboardingStatus.Pending, "Not pending");
        req.status = OnboardingStatus.Rejected;
        req.approvedBy = msg.sender;
        req.approvedAt = block.timestamp;
        emit SupplierOnboardingRejected(supplier, msg.sender, block.timestamp);
    }

    /// @notice Deactivate a supplier (SUPER_ADMIN only, also revokes all roles except DEFAULT_ADMIN_ROLE)
    /// @param supplier The address of the supplier to deactivate
    function deactivateSupplier(address supplier) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        require(isSupplier[supplier], "Not a supplier");
        isSupplier[supplier] = false;
        onboardingRequests[supplier].status = OnboardingStatus.Deactivated;
        // Revoke all roles except DEFAULT_ADMIN_ROLE
        bytes32[11] memory roles = [SUPER_ADMIN_ROLE, ADMIN_ROLE, MINER_ROLE, REFINER_ROLE, ALLOY_PRODUCER_ROLE, PRODUCT_MANUFACTURER_ROLE, DISTRIBUTOR_ROLE, SERVICE_PROVIDER_ROLE, RECYCLER_ROLE, AUDITOR_ROLE, REGULATOR_ROLE];
        for (uint i = 0; i < roles.length; i++) {
            if (hasRole(roles[i], supplier)) {
                _revokeRole(roles[i], supplier);
                emit RoleRevoked(supplier, roles[i], msg.sender, block.timestamp);
            }
        }
        emit SupplierDeactivated(supplier, msg.sender, block.timestamp);
    }

    /// @notice Revoke a role from an account (SUPER_ADMIN only)
    /// @param account The address to revoke the role from
    /// @param role The role to revoke
    function revokeRoleFrom(address account, bytes32 role) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        require(hasRole(role, account), "Account does not have role");
        _revokeRole(role, account);
        emit RoleRevoked(account, role, msg.sender, block.timestamp);
    }

    /// @notice Transfer the SUPER_ADMIN_ROLE to a new address (emergency recovery)
    /// @param newSuperAdmin The address to become the new super admin
    function transferSuperAdmin(address newSuperAdmin) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        require(newSuperAdmin != address(0), "Zero address");
        require(newSuperAdmin != msg.sender, "Cannot transfer to self");
        _grantRole(SUPER_ADMIN_ROLE, newSuperAdmin);
        _revokeRole(SUPER_ADMIN_ROLE, msg.sender);
        emit SuperAdminTransferred(msg.sender, newSuperAdmin, block.timestamp);
    }

    // --- Batch Processing for Large Scale Operations ---
    
    /// @notice Batch approve multiple suppliers at once
    /// @param suppliers Array of supplier addresses to approve
    /// @param roles Array of roles to grant (must match suppliers array length)
    function batchApproveSuppliers(
        address[] calldata suppliers,
        bytes32[] calldata roles
    ) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        require(suppliers.length == roles.length, "Arrays length mismatch");
        require(suppliers.length <= 50, "Max 50 suppliers per batch"); // Gas limit protection
        
        for (uint256 i = 0; i < suppliers.length; i++) {
            address supplier = suppliers[i];
            bytes32 role = roles[i];
            
            require(onboardingRequests[supplier].status == OnboardingStatus.Pending, "Supplier not pending");
            
            // Approve the supplier
            onboardingRequests[supplier].status = OnboardingStatus.Approved;
            onboardingRequests[supplier].approvedBy = msg.sender;
            onboardingRequests[supplier].approvedAt = block.timestamp;
            
            // Grant the requested role
            _grantRole(role, supplier);
            isSupplier[supplier] = true;
            
            emit SupplierOnboardingApproved(supplier, msg.sender, block.timestamp);
        }
    }
    
    /// @notice Batch reject multiple suppliers at once
    /// @param suppliers Array of supplier addresses to reject
    function batchRejectSuppliers(
        address[] calldata suppliers
    ) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        require(suppliers.length <= 50, "Max 50 suppliers per batch");
        
        for (uint256 i = 0; i < suppliers.length; i++) {
            address supplier = suppliers[i];
            
            require(onboardingRequests[supplier].status == OnboardingStatus.Pending, "Supplier not pending");
            
            onboardingRequests[supplier].status = OnboardingStatus.Rejected;
            onboardingRequests[supplier].approvedBy = msg.sender;
            onboardingRequests[supplier].approvedAt = block.timestamp;
            
            emit SupplierOnboardingRejected(supplier, msg.sender, block.timestamp);
        }
    }
    
    /// @notice Get pending suppliers for batch processing
    /// @param startIndex Starting index for pagination
    /// @param count Number of suppliers to return
    /// @return suppliers Array of pending supplier addresses
    /// @return totalPending Total number of pending suppliers
    function getPendingSuppliers(
        uint256 startIndex,
        uint256 count
    ) external view returns (address[] memory suppliers, uint256 totalPending) {
        // This would require additional storage to track pending suppliers efficiently
        // For now, returning empty arrays - implementation would need supplier list tracking
        suppliers = new address[](0);
        totalPending = 0;
    }
    
    /// @notice Bulk supplier status update
    /// @param suppliers Array of supplier addresses
    /// @param statuses Array of new statuses
    function bulkUpdateSupplierStatus(
        address[] calldata suppliers,
        OnboardingStatus[] calldata statuses
    ) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        require(suppliers.length == statuses.length, "Arrays length mismatch");
        require(suppliers.length <= 50, "Max 50 suppliers per batch");
        
        for (uint256 i = 0; i < suppliers.length; i++) {
            address supplier = suppliers[i];
            OnboardingStatus status = statuses[i];
            
            onboardingRequests[supplier].status = status;
            onboardingRequests[supplier].approvedBy = msg.sender;
            onboardingRequests[supplier].approvedAt = block.timestamp;
            
            if (status == OnboardingStatus.Approved) {
                isSupplier[supplier] = true;
            } else if (status == OnboardingStatus.Deactivated) {
                isSupplier[supplier] = false;
            }
        }
    }

    // --- Supplier Management & Analytics ---
    
    /// @notice Get supplier statistics for dashboard
    /// @return totalSuppliers Total number of approved suppliers
    /// @return pendingSuppliers Number of pending approvals
    /// @return activeSuppliers Number of active suppliers
    /// @return deactivatedSuppliers Number of deactivated suppliers
    function getSupplierStats() external view returns (
        uint256 totalSuppliers,
        uint256 pendingSuppliers,
        uint256 activeSuppliers,
        uint256 deactivatedSuppliers
    ) {
        // Note: This is a simplified implementation
        // Full implementation would require additional storage tracking
        totalSuppliers = 0;
        pendingSuppliers = 0;
        activeSuppliers = 0;
        deactivatedSuppliers = 0;
    }
    
    /// @notice Get suppliers by role for management
    /// @param role Role to filter by
    /// @param startIndex Starting index for pagination
    /// @param count Number of suppliers to return
    /// @return suppliers Array of supplier addresses with the role
    function getSuppliersByRole(
        bytes32 role,
        uint256 startIndex,
        uint256 count
    ) external view returns (address[] memory suppliers) {
        // Note: This would require additional storage for efficient role-based queries
        suppliers = new address[](0);
    }
    
    /// @notice Bulk deactivate suppliers
    /// @param suppliers Array of supplier addresses to deactivate
    function bulkDeactivateSuppliers(
        address[] calldata suppliers
    ) external onlyRole(SUPER_ADMIN_ROLE) whenNotPaused {
        require(suppliers.length <= 50, "Max 50 suppliers per batch");
        
        for (uint256 i = 0; i < suppliers.length; i++) {
            address supplier = suppliers[i];
            
            if (isSupplier[supplier]) {
                isSupplier[supplier] = false;
                onboardingRequests[supplier].status = OnboardingStatus.Deactivated;
                onboardingRequests[supplier].approvedBy = msg.sender;
                onboardingRequests[supplier].approvedAt = block.timestamp;
                
                emit SupplierDeactivated(supplier, msg.sender, block.timestamp);
            }
        }
    }
    
    /// @notice Check if supplier is active and has specific role
    /// @param supplier Supplier address to check
    /// @param role Role to verify
    /// @return isActive Whether supplier is active
    /// @return hasRole Whether supplier has the specified role
    function getSupplierStatus(
        address supplier,
        bytes32 role
    ) external view returns (bool isActive, bool hasRole) {
        isActive = isSupplier[supplier];
        hasRole = this.hasRole(role, supplier);
    }
    
    /// @notice Get all roles for a supplier
    /// @param supplier Supplier address
    /// @return roles Array of role hashes the supplier has
    function getSupplierRoles(address supplier) external view returns (bytes32[] memory roles) {
        // Note: This would require additional storage to track all roles efficiently
        roles = new bytes32[](0);
    }

    // --- Passport Management ---
    /// @notice Create a new aluminium passport
    function createPassport(
        string memory passportId,
        string memory origin,
        string memory manufacturer,
        string memory alloyComposition,
        string memory certifier,
        string memory ipfsHash,
        uint256 esgScore,
        uint256 recycledContent
    ) external onlyRoleOrAdmin(PRODUCT_MANUFACTURER_ROLE) whenNotPaused {
        require(passports[passportId].createdAt == 0, "Passport exists");
        require(bytes(passportId).length > 0, "passportId required");
        require(bytes(origin).length > 0, "origin required");
        require(bytes(manufacturer).length > 0, "manufacturer required");
        require(bytes(alloyComposition).length > 0, "alloy required");
        require(bytes(certifier).length > 0, "certifier required");
        require(bytes(ipfsHash).length > 0, "ipfsHash required");
        require(_isValidIPFSHash(ipfsHash), "Invalid IPFS hash");
        Passport storage p = passports[passportId];
        p.passportId = passportId;
        p.origin = origin;
        p.manufacturer = manufacturer;
        p.alloyComposition = alloyComposition;
        p.certifier = certifier;
        p.ipfsHash = ipfsHash;
        p.esgScore = esgScore;
        p.recycledContent = recycledContent;
        p.createdBy = msg.sender;
        p.createdAt = block.timestamp;
        p.updatedAt = block.timestamp;
        p.isActive = true;
        allPassportIds.push(passportId);
        emit PassportCreated(passportId, msg.sender, manufacturer, origin, block.timestamp);
    }

    /// @notice Update an existing passport
    function updatePassport(
        string memory passportId,
        string memory ipfsHash,
        uint256 esgScore,
        uint256 recycledContent
    ) external onlyRoleOrAdmin(PRODUCT_MANUFACTURER_ROLE) whenNotPaused {
        Passport storage p = passports[passportId];
        require(p.createdAt != 0, "Not found");
        require(p.isActive, "Inactive");
        require(bytes(ipfsHash).length > 0, "ipfsHash required");
        require(_isValidIPFSHash(ipfsHash), "Invalid IPFS hash");
        p.ipfsHash = ipfsHash;
        p.esgScore = esgScore;
        p.recycledContent = recycledContent;
        p.updatedAt = block.timestamp;
        emit PassportUpdated(passportId, msg.sender, block.timestamp);
    }

    /// @notice Deactivate a passport (ADMIN or higher)
    function deactivatePassport(string memory passportId) external onlyRoleOrAdmin(ADMIN_ROLE) whenNotPaused {
        Passport storage p = passports[passportId];
        require(p.createdAt != 0, "Not found");
        require(p.isActive, "Already inactive");
        p.isActive = false;
        p.updatedAt = block.timestamp;
        emit PassportDeactivated(passportId, msg.sender, block.timestamp);
    }

    // --- Certifications & Supply Chain ---
    /// @notice Add a certification to a passport
    function addCertification(string memory passportId, string memory certification) external onlyRoleOrAdmin(AUDITOR_ROLE) whenNotPaused {
        Passport storage p = passports[passportId];
        require(p.createdAt != 0, "Not found");
        require(bytes(certification).length > 0, "certification required");
        require(p.certifications.length < 50, "Too many certifications");
        p.certifications.push(certification);
        p.updatedAt = block.timestamp;
        emit CertificationAdded(passportId, certification, msg.sender, block.timestamp);
    }

    /// @notice Add a supply chain step to a passport
    function addSupplyChainStep(string memory passportId, string memory step) external onlyRoleOrAdmin(PRODUCT_MANUFACTURER_ROLE) whenNotPaused {
        Passport storage p = passports[passportId];
        require(p.createdAt != 0, "Not found");
        require(bytes(step).length > 0, "step required");
        require(p.supplyChainSteps.length < 100, "Too many steps");
        p.supplyChainSteps.push(step);
        p.updatedAt = block.timestamp;
        emit SupplyChainStepAdded(passportId, step, msg.sender, block.timestamp);
    }

    // --- Getters ---
    /// @notice Get passport details
    function getPassport(string memory passportId) external view returns (
        string memory, string memory, string memory, string memory, string memory, uint256, uint256, bool, string memory, address, uint256, uint256
    ) {
        Passport storage p = passports[passportId];
        return (
            p.passportId,
            p.origin,
            p.manufacturer,
            p.alloyComposition,
            p.certifier,
            p.esgScore,
            p.recycledContent,
            p.isActive,
            p.ipfsHash,
            p.createdBy,
            p.createdAt,
            p.updatedAt
        );
    }

    /// @notice Get certifications for a passport
    function getCertifications(string memory passportId) external view returns (string[] memory) {
        return passports[passportId].certifications;
    }

    /// @notice Get supply chain steps for a passport
    function getSupplyChainSteps(string memory passportId) external view returns (string[] memory) {
        return passports[passportId].supplyChainSteps;
    }

    /// @notice Get all passport IDs
    function getAllPassportIds() external view returns (string[] memory) {
        return allPassportIds;
    }

    // --- Internal helpers ---
    function _stringToRole(string memory role) internal pure returns (bytes32) {
        if (keccak256(bytes(role)) == keccak256("SUPER_ADMIN_ROLE")) return SUPER_ADMIN_ROLE;
        if (keccak256(bytes(role)) == keccak256("ADMIN_ROLE")) return ADMIN_ROLE;
        if (keccak256(bytes(role)) == keccak256("AUDITOR_ROLE")) return AUDITOR_ROLE;
        if (keccak256(bytes(role)) == keccak256("MINER_ROLE")) return MINER_ROLE;
        if (keccak256(bytes(role)) == keccak256("REFINER_ROLE")) return REFINER_ROLE;
        if (keccak256(bytes(role)) == keccak256("ALLOY_PRODUCER_ROLE")) return ALLOY_PRODUCER_ROLE;
        if (keccak256(bytes(role)) == keccak256("PRODUCT_MANUFACTURER_ROLE")) return PRODUCT_MANUFACTURER_ROLE;
        if (keccak256(bytes(role)) == keccak256("DISTRIBUTOR_ROLE")) return DISTRIBUTOR_ROLE;
        if (keccak256(bytes(role)) == keccak256("SERVICE_PROVIDER_ROLE")) return SERVICE_PROVIDER_ROLE;
        if (keccak256(bytes(role)) == keccak256("MANUFACTURER_ROLE")) return PRODUCT_MANUFACTURER_ROLE; // Backward compatibility
        if (keccak256(bytes(role)) == keccak256("RECYCLER_ROLE")) return RECYCLER_ROLE;
        if (keccak256(bytes(role)) == keccak256("REGULATOR_ROLE")) return REGULATOR_ROLE;
        revert("Unknown role");
    }

    function _isValidIPFSHash(string memory ipfsHash) internal pure returns (bool) {
        // Basic check: must start with "Qm" and be 46 chars (CIDv0), or "b" and 59 chars (CIDv1 base32)
        bytes memory b = bytes(ipfsHash);
        if (b.length == 46 && b[0] == "Q" && b[1] == "m") return true;
        if (b.length == 59 && b[0] == "b") return true;
        return false;
    }

    // --- Storage gap for upgradeability ---
    uint256[50] private __gap;
}
