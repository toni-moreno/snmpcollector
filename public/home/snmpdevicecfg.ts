export class SnmpDeviceCfg {
    id: string;
    Host: string;
    Port: number;
    Retries: number;
    Timeout: number;
    SnmpVersion: string;
    Community: string;
    V3SecLevel: string;
    V3AuthUser: string;
    V3AuthPass: string;
    V3AuthProt: string;
    V3PrivPass: string;
    V3PrivProt: string;
    Freq: number;
    Config: string;
    LogLevel: string;
    LogFile: string;
    SnmpDebug: boolean;
    DeviceTagName: string;
    DeviceTagValue: string;
    Extratags: Array<string>;
    
    constructor(id: number, 
                description: string,
                dueDate: Date,
                complete: boolean) {
      this.id = 'hola';
      this.Host= 'localhost';
      this.Port = 161;
    }
    /*
    setComplete() {
      this.complete = true;
      this.completedDate = new Date();
    }*/
  }
