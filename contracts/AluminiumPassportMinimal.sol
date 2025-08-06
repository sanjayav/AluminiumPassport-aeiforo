// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";

contract AluminiumPassportMinimal is Initializable, UUPSUpgradeable, AccessControlUpgradeable {
    string public constant VERSION = "minimal-roles-1.0.0";
    bytes32 public constant SUPER_ADMIN_ROLE = keccak256("SUPER_ADMIN_ROLE");
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");

    struct Passport {
        string passportId;
        string origin;
        string manufacturer;
        string alloyComposition;
        string certifier;
        string ipfsHash;
        uint256 esgScore;
        uint256 recycledContent;
        bool isActive;
    }

    mapping(string => Passport) private passports;

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

    mapping(address => SupplierOnboarding) public onboardingRequests;

    event PassportCreated(string indexed passportId, address indexed createdBy, string manufacturer, string origin, uint256 timestamp);
    event SupplierOnboardingRequested(address indexed supplier, string roleRequested, string companyName, address indexed requestedBy, uint256 timestamp);

    function initialize(address superAdmin, address admin) public initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        _grantRole(DEFAULT_ADMIN_ROLE, superAdmin);
        _grantRole(SUPER_ADMIN_ROLE, superAdmin);
        _grantRole(ADMIN_ROLE, admin);
    }

    function _authorizeUpgrade(address) internal override {}

    function createPassport(
        string memory passportId,
        string memory origin,
        string memory manufacturer,
        string memory alloyComposition,
        string memory certifier,
        string memory ipfsHash,
        uint256 esgScore,
        uint256 recycledContent
    ) public onlyRoleOrAdmin(ADMIN_ROLE) {
        passports[passportId] = Passport({
            passportId: passportId,
            origin: origin,
            manufacturer: manufacturer,
            alloyComposition: alloyComposition,
            certifier: certifier,
            ipfsHash: ipfsHash,
            esgScore: esgScore,
            recycledContent: recycledContent,
            isActive: true
        });
        emit PassportCreated(passportId, msg.sender, manufacturer, origin, block.timestamp);
    }

    function getPassport(string memory passportId) public view returns (
        string memory, string memory, string memory, string memory, string memory, string memory, uint256, uint256, bool
    ) {
        Passport storage p = passports[passportId];
        return (
            p.passportId,
            p.origin,
            p.manufacturer,
            p.alloyComposition,
            p.certifier,
            p.ipfsHash,
            p.esgScore,
            p.recycledContent,
            p.isActive
        );
    }

    function requestSupplierOnboarding(
        string memory roleRequested,
        string memory companyName,
        string memory metadataIPFS
    ) public onlyRole(SUPER_ADMIN_ROLE) {
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

    modifier onlyRoleOrAdmin(bytes32 role) {
        require(
            hasRole(role, msg.sender) || hasRole(ADMIN_ROLE, msg.sender) || hasRole(SUPER_ADMIN_ROLE, msg.sender),
            "Not authorized"
        );
        _;
    }
}