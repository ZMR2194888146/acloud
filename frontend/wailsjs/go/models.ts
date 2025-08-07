export namespace main {
	
	export class AuthResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    size: number;
	    isDir: boolean;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.isDir = source["isDir"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FilePreview {
	    mimeType: string;
	    content: string;
	    isBase64: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FilePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mimeType = source["mimeType"];
	        this.content = source["content"];
	        this.isBase64 = source["isBase64"];
	    }
	}
	export class ISCSIConnection {
	    id: string;
	    targetIqn: string;
	    portal: string;
	    status: string;
	    // Go type: time
	    connectedAt: any;
	    // Go type: time
	    lastActivity: any;
	    bytesRead: number;
	    bytesWritten: number;
	    iops: number;
	    latency: number;
	    bandwidth: number;
	
	    static createFrom(source: any = {}) {
	        return new ISCSIConnection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.targetIqn = source["targetIqn"];
	        this.portal = source["portal"];
	        this.status = source["status"];
	        this.connectedAt = this.convertValues(source["connectedAt"], null);
	        this.lastActivity = this.convertValues(source["lastActivity"], null);
	        this.bytesRead = source["bytesRead"];
	        this.bytesWritten = source["bytesWritten"];
	        this.iops = source["iops"];
	        this.latency = source["latency"];
	        this.bandwidth = source["bandwidth"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ISCSIDiscoveredTarget {
	    iqn: string;
	    portal: string;
	    targetName: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ISCSIDiscoveredTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.iqn = source["iqn"];
	        this.portal = source["portal"];
	        this.targetName = source["targetName"];
	        this.status = source["status"];
	    }
	}
	export class ISCSIDisk {
	    devicePath: string;
	    targetIqn: string;
	    lun: number;
	    size: number;
	    model: string;
	    serial: string;
	    status: string;
	    mountPoint: string;
	    fileSystem: string;
	    usedSpace: number;
	    freeSpace: number;
	
	    static createFrom(source: any = {}) {
	        return new ISCSIDisk(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.devicePath = source["devicePath"];
	        this.targetIqn = source["targetIqn"];
	        this.lun = source["lun"];
	        this.size = source["size"];
	        this.model = source["model"];
	        this.serial = source["serial"];
	        this.status = source["status"];
	        this.mountPoint = source["mountPoint"];
	        this.fileSystem = source["fileSystem"];
	        this.usedSpace = source["usedSpace"];
	        this.freeSpace = source["freeSpace"];
	    }
	}
	export class ISCSIInitiatorConfig {
	    initiatorName: string;
	    defaultPort: number;
	    loginTimeout: number;
	    logoutTimeout: number;
	    nopOutInterval: number;
	    nopOutTimeout: number;
	    maxConnections: number;
	    headerDigest: string;
	    dataDigest: string;
	    maxRecvDataSegmentLength: number;
	
	    static createFrom(source: any = {}) {
	        return new ISCSIInitiatorConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.initiatorName = source["initiatorName"];
	        this.defaultPort = source["defaultPort"];
	        this.loginTimeout = source["loginTimeout"];
	        this.logoutTimeout = source["logoutTimeout"];
	        this.nopOutInterval = source["nopOutInterval"];
	        this.nopOutTimeout = source["nopOutTimeout"];
	        this.maxConnections = source["maxConnections"];
	        this.headerDigest = source["headerDigest"];
	        this.dataDigest = source["dataDigest"];
	        this.maxRecvDataSegmentLength = source["maxRecvDataSegmentLength"];
	    }
	}
	export class ISCSIPerformanceStats {
	    totalConnections: number;
	    activeConnections: number;
	    totalDisks: number;
	    mountedDisks: number;
	    totalIops: number;
	    totalBandwidth: number;
	    averageLatency: number;
	    connectionStats: ISCSIConnection[];
	    diskStats: ISCSIDisk[];
	
	    static createFrom(source: any = {}) {
	        return new ISCSIPerformanceStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalConnections = source["totalConnections"];
	        this.activeConnections = source["activeConnections"];
	        this.totalDisks = source["totalDisks"];
	        this.mountedDisks = source["mountedDisks"];
	        this.totalIops = source["totalIops"];
	        this.totalBandwidth = source["totalBandwidth"];
	        this.averageLatency = source["averageLatency"];
	        this.connectionStats = this.convertValues(source["connectionStats"], ISCSIConnection);
	        this.diskStats = this.convertValues(source["diskStats"], ISCSIDisk);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MinioConfig {
	    endpoint: string;
	    accessKeyId: string;
	    secretAccessKey: string;
	    useSSL: boolean;
	    bucketName: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MinioConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpoint = source["endpoint"];
	        this.accessKeyId = source["accessKeyId"];
	        this.secretAccessKey = source["secretAccessKey"];
	        this.useSSL = source["useSSL"];
	        this.bucketName = source["bucketName"];
	        this.enabled = source["enabled"];
	    }
	}
	export class MinioFileInfo {
	    name: string;
	    path: string;
	    size: number;
	    // Go type: time
	    lastModified: any;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MinioFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.lastModified = this.convertValues(source["lastModified"], null);
	        this.isDir = source["isDir"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TrayMenuItem {
	    id: string;
	    label: string;
	    type: string;
	    enabled: boolean;
	    checked: boolean;
	    shortcut: string;
	
	    static createFrom(source: any = {}) {
	        return new TrayMenuItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.checked = source["checked"];
	        this.shortcut = source["shortcut"];
	    }
	}

}

