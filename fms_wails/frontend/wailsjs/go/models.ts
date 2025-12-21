export namespace model {
	
	export class Config {
	    connectionMode: string;
	    agentServerURL: string;
	    timeoutSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionMode = source["connectionMode"];
	        this.agentServerURL = source["agentServerURL"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	    }
	}
	export class RuleResult {
	    rule: string;
	    text: string;
	    status: string;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new RuleResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rule = source["rule"];
	        this.text = source["text"];
	        this.status = source["status"];
	        this.reason = source["reason"];
	    }
	}
	export class DeployHistory {
	    id: number;
	    // Go type: time
	    timestamp: any;
	    deviceIp: string;
	    templateVersion: string;
	    status: string;
	    results: RuleResult[];
	
	    static createFrom(source: any = {}) {
	        return new DeployHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.deviceIp = source["deviceIp"];
	        this.templateVersion = source["templateVersion"];
	        this.status = source["status"];
	        this.results = this.convertValues(source["results"], RuleResult);
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
	export class Firewall {
	    index: number;
	    deviceName: string;
	    serverStatus: string;
	    deployStatus: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new Firewall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.deviceName = source["deviceName"];
	        this.serverStatus = source["serverStatus"];
	        this.deployStatus = source["deployStatus"];
	        this.version = source["version"];
	    }
	}
	
	export class Template {
	    version: string;
	    contents: string;
	
	    static createFrom(source: any = {}) {
	        return new Template(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.contents = source["contents"];
	    }
	}

}

