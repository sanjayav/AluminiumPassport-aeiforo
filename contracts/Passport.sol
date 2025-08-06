// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

contract AluminiumPassport is AccessControl, ReentrancyGuard {
    using Counters for Counters.Counter;
    
    bytes32 public constant MINER_ROLE = keccak256("MINER_ROLE");
    bytes32 public constant RECYCLER_ROLE = keccak256("RECYCLER_ROLE");
    bytes32 public constant CERTIFIER_ROLE = keccak256("CERTIFIER_ROLE");
    bytes32 public constant MANUFACTURER_ROLE = keccak256("MANUFACTURER_ROLE");
    
    Counters.Counter private _passportIds;
    
    struct Passport {
        string passportId;
        string manufacturer;
        string origin;
        string bauxiteSource;
        string alloyComposition;
        uint256 recycledContent;      // Percentage (0-100)
        uint256 esgScore;            // ESG score (0-100)
        uint256 co2Footprint;        // CO2 footprint in kg
        string certifier;
        string ipfsHash;
        uint256 createdAt;
        uint256 updatedAt;
        bool isActive;
        
        // Manufacturing details
        string processType;
        uint256 energyUsed;          // kWh
        string energySource;
        uint256 waterUsed;           // Liters
        uint256 wasteGenerated;      // kg
        
        // Supply chain tracking
        string[] supplyChainSteps;
        uint256[] timestamps;
        
        // Compliance & Certifications
        string[] certifications;
        string complianceStandards;
        uint256 certificationDate;
        
        // Recycling history
        uint256 timesRecycled;
        string lastRecyclingMethod;
        uint256 lastRecyclingDate;
    }
    
    struct ESGMetrics {
        uint256 environmentalScore;  // 0-100
        uint256 socialScore;         // 0-100
        uint256 governanceScore;     // 0-100
        uint256 overallScore;        // 0-100
        uint256 lastUpdated;
    }
    
    mapping(string => Passport) public passports;
    mapping(string => ESGMetrics) public esgMetrics;
    mapping(string => bool) public passportExists;
    mapping(address => string[]) public userPassports;
    
    string[] public allPassportIds;
    
    event PassportRegistered(
        string indexed passportId,
        address indexed registeredBy,
        string manufacturer,
        string origin,
        uint256 timestamp
    );
    
    event PassportUpdated(
        string indexed passportId,
        address indexed updatedBy,
        string fieldUpdated,
        uint256 timestamp
    );
    
    event RecycledContentUpdated(
        string indexed passportId,
        uint256 oldPercentage,
        uint256 newPercentage,
        address indexed updatedBy
    );
    
    event ESGScoreUpdated(
        string indexed passportId,
        uint256 newScore,
        address indexed updatedBy
    );
    
    event SupplyChainStepAdded(
        string indexed passportId,
        string step,
        uint256 timestamp,
        address indexed addedBy
    );
    
    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(MINER_ROLE, msg.sender);
        _grantRole(RECYCLER_ROLE, msg.sender);
        _grantRole(CERTIFIER_ROLE, msg.sender);
        _grantRole(MANUFACTURER_ROLE, msg.sender);
    }
    
    modifier passportExistsModifier(string memory _passportId) {
        require(passportExists[_passportId], "Passport does not exist");
        _;
    }
    
    function registerPassport(
        string memory _passportId,
        string memory _manufacturer,
        string memory _origin,
        string memory _bauxiteSource,
        string memory _alloyComposition,
        uint256 _recycledContent,
        uint256 _co2Footprint,
        string memory _certifier,
        string memory _ipfsHash,
        string memory _processType,
        uint256 _energyUsed,
        string memory _energySource
    ) external onlyRole(MINER_ROLE) nonReentrant {
        require(!passportExists[_passportId], "Passport already exists");
        require(_recycledContent <= 100, "Recycled content cannot exceed 100%");
        require(bytes(_passportId).length > 0, "Passport ID cannot be empty");
        require(bytes(_manufacturer).length > 0, "Manufacturer cannot be empty");
        
        _passportIds.increment();
        
        Passport storage newPassport = passports[_passportId];
        newPassport.passportId = _passportId;
        newPassport.manufacturer = _manufacturer;
        newPassport.origin = _origin;
        newPassport.bauxiteSource = _bauxiteSource;
        newPassport.alloyComposition = _alloyComposition;
        newPassport.recycledContent = _recycledContent;
        newPassport.co2Footprint = _co2Footprint;
        newPassport.certifier = _certifier;
        newPassport.ipfsHash = _ipfsHash;
        newPassport.processType = _processType;
        newPassport.energyUsed = _energyUsed;
        newPassport.energySource = _energySource;
        newPassport.createdAt = block.timestamp;
        newPassport.updatedAt = block.timestamp;
        newPassport.isActive = true;
        newPassport.timesRecycled = 0;
        
        // Initialize supply chain with first step
        newPassport.supplyChainSteps.push("Mining/Extraction");
        newPassport.timestamps.push(block.timestamp);
        
        passportExists[_passportId] = true;
        userPassports[msg.sender].push(_passportId);
        allPassportIds.push(_passportId);
        
        emit PassportRegistered(_passportId, msg.sender, _manufacturer, _origin, block.timestamp);
    }
    
    function updateRecycledContent(
        string memory _passportId,
        uint256 _newPercentage,
        string memory _recyclingMethod
    ) external onlyRole(RECYCLER_ROLE) passportExistsModifier(_passportId) {
        require(_newPercentage <= 100, "Recycled content cannot exceed 100%");
        
        Passport storage passport = passports[_passportId];
        uint256 oldPercentage = passport.recycledContent;
        
        passport.recycledContent = _newPercentage;
        passport.timesRecycled += 1;
        passport.lastRecyclingMethod = _recyclingMethod;
        passport.lastRecyclingDate = block.timestamp;
        passport.updatedAt = block.timestamp;
        
        // Add to supply chain
        passport.supplyChainSteps.push("Recycling");
        passport.timestamps.push(block.timestamp);
        
        emit RecycledContentUpdated(_passportId, oldPercentage, _newPercentage, msg.sender);
        emit PassportUpdated(_passportId, msg.sender, "recycledContent", block.timestamp);
    }
    
    function updateESGScore(
        string memory _passportId,
        uint256 _environmentalScore,
        uint256 _socialScore,
        uint256 _governanceScore
    ) external onlyRole(CERTIFIER_ROLE) passportExistsModifier(_passportId) {
        require(_environmentalScore <= 100 && _socialScore <= 100 && _governanceScore <= 100, 
                "Scores must be between 0-100");
        
        uint256 overallScore = (_environmentalScore + _socialScore + _governanceScore) / 3;
        
        ESGMetrics storage metrics = esgMetrics[_passportId];
        metrics.environmentalScore = _environmentalScore;
        metrics.socialScore = _socialScore;
        metrics.governanceScore = _governanceScore;
        metrics.overallScore = overallScore;
        metrics.lastUpdated = block.timestamp;
        
        passports[_passportId].esgScore = overallScore;
        passports[_passportId].updatedAt = block.timestamp;
        
        emit ESGScoreUpdated(_passportId, overallScore, msg.sender);
        emit PassportUpdated(_passportId, msg.sender, "esgScore", block.timestamp);
    }
    
    function addSupplyChainStep(
        string memory _passportId,
        string memory _step
    ) external passportExistsModifier(_passportId) {
        require(hasRole(MANUFACTURER_ROLE, msg.sender) || 
                hasRole(RECYCLER_ROLE, msg.sender) || 
                hasRole(CERTIFIER_ROLE, msg.sender), 
                "Unauthorized to add supply chain step");
        
        Passport storage passport = passports[_passportId];
        passport.supplyChainSteps.push(_step);
        passport.timestamps.push(block.timestamp);
        passport.updatedAt = block.timestamp;
        
        emit SupplyChainStepAdded(_passportId, _step, block.timestamp, msg.sender);
        emit PassportUpdated(_passportId, msg.sender, "supplyChain", block.timestamp);
    }
    
    function addCertification(
        string memory _passportId,
        string memory _certification
    ) external onlyRole(CERTIFIER_ROLE) passportExistsModifier(_passportId) {
        Passport storage passport = passports[_passportId];
        passport.certifications.push(_certification);
        passport.certificationDate = block.timestamp;
        passport.updatedAt = block.timestamp;
        
        emit PassportUpdated(_passportId, msg.sender, "certification", block.timestamp);
    }
    
    function updateIPFSHash(
        string memory _passportId,
        string memory _newHash
    ) external passportExistsModifier(_passportId) {
        require(hasRole(MINER_ROLE, msg.sender) || 
                hasRole(MANUFACTURER_ROLE, msg.sender) || 
                hasRole(CERTIFIER_ROLE, msg.sender), 
                "Unauthorized to update IPFS hash");
        
        passports[_passportId].ipfsHash = _newHash;
        passports[_passportId].updatedAt = block.timestamp;
        
        emit PassportUpdated(_passportId, msg.sender, "ipfsHash", block.timestamp);
    }
    
    function deactivatePassport(
        string memory _passportId
    ) external onlyRole(DEFAULT_ADMIN_ROLE) passportExistsModifier(_passportId) {
        passports[_passportId].isActive = false;
        passports[_passportId].updatedAt = block.timestamp;
        
        emit PassportUpdated(_passportId, msg.sender, "deactivated", block.timestamp);
    }
    
    // View functions
    function getPassportDetails(string memory _passportId) 
        external 
        view 
        passportExistsModifier(_passportId) 
        returns (Passport memory) {
        return passports[_passportId];
    }
    
    function getESGMetrics(string memory _passportId) 
        external 
        view 
        passportExistsModifier(_passportId) 
        returns (ESGMetrics memory) {
        return esgMetrics[_passportId];
    }
    
    function getSupplyChain(string memory _passportId) 
        external 
        view 
        passportExistsModifier(_passportId) 
        returns (string[] memory steps, uint256[] memory timestamps) {
        Passport storage passport = passports[_passportId];
        return (passport.supplyChainSteps, passport.timestamps);
    }
    
    function getCertifications(string memory _passportId) 
        external 
        view 
        passportExistsModifier(_passportId) 
        returns (string[] memory) {
        return passports[_passportId].certifications;
    }
    
    function getUserPassports(address _user) external view returns (string[] memory) {
        return userPassports[_user];
    }
    
    function getAllPassportIds() external view returns (string[] memory) {
        return allPassportIds;
    }
    
    function getTotalPassports() external view returns (uint256) {
        return _passportIds.current();
    }
    
    function isPassportActive(string memory _passportId) 
        external 
        view 
        passportExistsModifier(_passportId) 
        returns (bool) {
        return passports[_passportId].isActive;
    }
    
    // Role management functions
    function grantMinerRole(address _account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        grantRole(MINER_ROLE, _account);
    }
    
    function grantRecyclerRole(address _account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        grantRole(RECYCLER_ROLE, _account);
    }
    
    function grantCertifierRole(address _account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        grantRole(CERTIFIER_ROLE, _account);
    }
    
    function grantManufacturerRole(address _account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        grantRole(MANUFACTURER_ROLE, _account);
    }
}