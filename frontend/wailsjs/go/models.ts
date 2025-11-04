export namespace services {
	
	export class GeneralConfig {
	    autoStart: boolean;
	    checkInterval: number;
	    switchDelay: number;
	    enableLogging: boolean;
	    logLevel: string;
	    showNotifications: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GeneralConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoStart = source["autoStart"];
	        this.checkInterval = source["checkInterval"];
	        this.switchDelay = source["switchDelay"];
	        this.enableLogging = source["enableLogging"];
	        this.logLevel = source["logLevel"];
	        this.showNotifications = source["showNotifications"];
	    }
	}
	export class Rule {
	    app: string;
	    window: string;
	    input: string;
	    enabled: boolean;
	    priority: number;
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app = source["app"];
	        this.window = source["window"];
	        this.input = source["input"];
	        this.enabled = source["enabled"];
	        this.priority = source["priority"];
	    }
	}
	export class Config {
	    rules: Rule[];
	    general: GeneralConfig;
	    lastModified: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rules = this.convertValues(source["rules"], Rule);
	        this.general = this.convertValues(source["general"], GeneralConfig);
	        this.lastModified = source["lastModified"];
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
	
	export class InputMethod {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new InputMethod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class LogEntry {
	    // Go type: time
	    timestamp: any;
	    level: string;
	    message: string;
	    appName?: string;
	    input?: string;
	    action?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.appName = source["appName"];
	        this.input = source["input"];
	        this.action = source["action"];
	        this.error = source["error"];
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
	
	export class WindowInfo {
	    appName: string;
	    appPath: string;
	    windowName: string;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.appPath = source["appPath"];
	        this.windowName = source["windowName"];
	        this.pid = source["pid"];
	    }
	}

}

