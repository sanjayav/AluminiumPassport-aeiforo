
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract AluminiumPassport {
    struct Passport {
        string passportId;
        string bauxiteOrigin;
        string alloyComposition;
        string certificationAgency;
        string verifierSignature;
    }

    mapping(string => Passport) private passports;
    address public owner;

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Not authorized");
        _;
    }

    function createPassport(
        string memory _passportId,
        string memory _bauxiteOrigin,
        string memory _alloyComposition,
        string memory _certificationAgency,
        string memory _verifierSignature
    ) public onlyOwner {
        passports[_passportId] = Passport(
            _passportId,
            _bauxiteOrigin,
            _alloyComposition,
            _certificationAgency,
            _verifierSignature
        );
    }

    function getPassport(string memory _passportId) public view returns (
        string memory,
        string memory,
        string memory,
        string memory,
        string memory
    ) {
        Passport memory p = passports[_passportId];
        return (
            p.passportId,
            p.bauxiteOrigin,
            p.alloyComposition,
            p.certificationAgency,
            p.verifierSignature
        );
    }
}
