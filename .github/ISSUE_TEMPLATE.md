Read before posting:

- Please search here on GitHub for similar issues before creating a new issue.
- Please review Documentation https://github.com/toni-moreno/snmpcollector/wiki before creating a new issue.
- Checkout How to troubleshoot issue the troubleshoot guide: https://github.com/toni-moreno/snmpcollector/wiki/Troubleshooting

For faster fixes and/or improvements would be nice if we can work with a simulation snapshot, plese follow this commands and attach the `sim_data_<DATE>_#<issue>.tar.gz` file to the issue.

```
# git clone https://github.com/etingof/snmpsim.git`
# cd snmpsim
# sudo python3 setup.py install
# mkdir sim_data
# snmpsim-record-commands --protocol-version 2c --community <your_comunity> --agent-udpv4-endpoint=<your_device>:161 --output-file=./sim_data/device_data
# tar zcvf sim_data_<DATE>_#<issue>.tar.gz sim_data
```

Please prefix your title with [Bug] or [Feature request].

Please include this information:
- What SnmpCollector version are you using?
- What OS are you running snmpcollector on?
- What did you do?
- What was the expected result?
- What happened instead?
- If related with process panics.
  - Include the related log files indicated in the troubleshooting guide.(snmpcollector.log `<deviceid>.log`)
- If related with configuration issues.
  - Include all  related configuration data ( and export file, a webui screenshot or writted by hand if you prefer)
- If webui seems to be freezed , :
  - Reload in another window de webui and, include raw network request & response: get by opening Chrome Dev Tools (F12, Ctrl+Shift+I on windows, Cmd+Opt+I on Mac), go the network tab.Review browser console for errors and also Include them.
