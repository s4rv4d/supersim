// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";

import {Unauthorized} from "@contracts-bedrock/libraries/errors/CommonErrors.sol";
import {Predeploys} from "@contracts-bedrock/libraries/Predeploys.sol";
import {SafeCall} from "@contracts-bedrock//libraries/SafeCall.sol";
import {IL2ToL2CrossDomainMessenger} from "@contracts-bedrock/L2/interfaces/IL2ToL2CrossDomainMessenger.sol";

contract SToken is Initializable, ERC20Upgradeable, OwnableUpgradeable, UUPSUpgradeable {

    address internal constant _MESSENGER = Predeploys.L2_TO_L2_CROSS_DOMAIN_MESSENGER;

    modifier onlyValidCaller() {
        require(_msgSender() == owner() || _msgSender() == _MESSENGER, "SToken: caller is not owner or messenger");
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize() initializer public {
        __ERC20_init("SToken", "STK");
        __Ownable_init();
        __UUPSUpgradeable_init();

        _mint(msg.sender, 1000 * 10 ** 18);
    }

    function _authorizeUpgrade(address newImplementation)
        internal
        onlyOwner
        override
    {}

    function transferOwnership(address newOwner) public override onlyValidCaller() {

    }
}