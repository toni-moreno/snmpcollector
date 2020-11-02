# Mock SnmpServer

This SnmpServer only implements server for unit test purposes, it takes part of the code from https://github.com/slayercat/GoSNMPServer, the main goal is simulate response data from snmp server for testing the SnmpCollector measurement types and any kind of metric post processing with known data. 

Minimal features for data querying simulation.

 * No Authentication.
 * No Snmp Bulk control parametrizations
 * No Snmp V3 testing
 * Supported GoSNMP query methods:
    * Get()
    * Walk()
    * BulkWalk()
