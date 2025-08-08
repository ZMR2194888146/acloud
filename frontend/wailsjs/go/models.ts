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
	export class ConflictFile {
	    path: string;
	    // Go type: time
	    localModTime: any;
	    // Go type: time
	    remoteModTime: any;
	    resolution: string;
	
	    static createFrom(source: any = {}) {
	        return new ConflictFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.localModTime = this.convertValues(source["localModTime"], null);
	        this.remoteModTime = this.convertValues(source["remoteModTime"], null);
	        this.resolution = source["resolution"];
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
	export class MinioConfig {
	    endpoint: string;
	    accessKeyID: string;
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
	        this.accessKeyID = source["accessKeyID"];
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
	export class SyncHistoryEntry {
	    id: string;
	    // Go type: time
	    timestamp: any;
	    syncMode: string;
	    filesUploaded: number;
	    filesDownloaded: number;
	    conflictCount: number;
	    errorCount: number;
	    duration: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncHistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.syncMode = source["syncMode"];
	        this.filesUploaded = source["filesUploaded"];
	        this.filesDownloaded = source["filesDownloaded"];
	        this.conflictCount = source["conflictCount"];
	        this.errorCount = source["errorCount"];
	        this.duration = source["duration"];
	        this.status = source["status"];
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
	export class SyncLogEntry {
	    // Go type: time
	    timestamp: any;
	    level: string;
	    message: string;
	    file: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.file = source["file"];
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
	export class SyncProgress {
	    totalFiles: number;
	    processedFiles: number;
	    uploadedFiles: number;
	    downloadedFiles: number;
	    currentFile: string;
	    progress: number;
	    status: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalFiles = source["totalFiles"];
	        this.processedFiles = source["processedFiles"];
	        this.uploadedFiles = source["uploadedFiles"];
	        this.downloadedFiles = source["downloadedFiles"];
	        this.currentFile = source["currentFile"];
	        this.progress = source["progress"];
	        this.status = source["status"];
	        this.error = source["error"];
	    }
	}
	export class SyncRule {
	    id: string;
	    name: string;
	    localPath: string;
	    remotePath: string;
	    direction: string;
	    filters: string[];
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SyncRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.localPath = source["localPath"];
	        this.remotePath = source["remotePath"];
	        this.direction = source["direction"];
	        this.filters = source["filters"];
	        this.enabled = source["enabled"];
	    }
	}
	export class SyncStatus {
	    running: boolean;
	    // Go type: time
	    lastSync: any;
	    filesUploaded: number;
	    filesDownloaded: number;
	    errors: string[];
	    syncMode: string;
	    conflictCount: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.lastSync = this.convertValues(source["lastSync"], null);
	        this.filesUploaded = source["filesUploaded"];
	        this.filesDownloaded = source["filesDownloaded"];
	        this.errors = source["errors"];
	        this.syncMode = source["syncMode"];
	        this.conflictCount = source["conflictCount"];
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
	    checked: boolean;
	    disabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TrayMenuItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.checked = source["checked"];
	        this.disabled = source["disabled"];
	    }
	}

}

