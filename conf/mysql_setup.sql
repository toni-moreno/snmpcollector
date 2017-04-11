create database snmpcollector;
GRANT USAGE ON `snmpcollector`.* to 'snmpcoluser'@'localhost' identified by 'snmpcolpass';
GRANT ALL PRIVILEGES ON `snmpcollector`.* to 'snmpcoluser'@'localhost' with grant option;
flush privileges;
