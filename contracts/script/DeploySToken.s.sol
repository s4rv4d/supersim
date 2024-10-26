// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import {Script, console} from "forge-std/Script.sol";
import {SToken} from "../src/SToken.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract DeploySToken is Script {
    function run() external {

        // Start broadcasting transactions
        vm.startBroadcast();

         // Deploy the implementation contract
        SToken sTokenImplementation = new SToken();

        // Prepare the initializer data
        bytes memory initializer = abi.encodeWithSelector(
            SToken.initialize.selector
        );

        // Deploy the proxy contract pointing to the implementation
        ERC1967Proxy proxy = new ERC1967Proxy(
            address(sTokenImplementation),
            initializer
        );

        // Cast the proxy to the SToken interface
        SToken sToken = SToken(address(proxy));

        console.log("SToken Proxy deployed at:", address(proxy));
        console.log("SToken Implementation deployed at:", address(sTokenImplementation));

        // Stop broadcasting transactions
        vm.stopBroadcast();
    }
}