export namespace models {
	
	export class Favorite {
	    id?: string;
	    toolId: string;
	    kind?: string;
	    name?: string;
	    payload?: Record<string, any>;
	    addedAt?: string;
	
	    static createFrom(source: any = {}) {
	        return new Favorite(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.toolId = source["toolId"];
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.payload = source["payload"];
	        this.addedAt = source["addedAt"];
	    }
	}
	export class History {
	    id?: string;
	    toolId: string;
	    usedAt: string;
	    input?: Record<string, any>;
	    output?: any;
	
	    static createFrom(source: any = {}) {
	        return new History(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.toolId = source["toolId"];
	        this.usedAt = source["usedAt"];
	        this.input = source["input"];
	        this.output = source["output"];
	    }
	}
	export class Settings {
	    theme: string;
	    language: string;
	    storagePath: string;
	    jsonIndent: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.language = source["language"];
	        this.storagePath = source["storagePath"];
	        this.jsonIndent = source["jsonIndent"];
	    }
	}
	export class ToolCapabilities {
	    fileInput: boolean;
	    fileOutput: boolean;
	    network: boolean;
	    streaming: boolean;
	    stateful: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolCapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileInput = source["fileInput"];
	        this.fileOutput = source["fileOutput"];
	        this.network = source["network"];
	        this.streaming = source["streaming"];
	        this.stateful = source["stateful"];
	    }
	}
	export class Tool {
	    id: string;
	    name: string;
	    description: string;
	    category: string;
	    icon: string;
	    version: string;
	    capabilities: ToolCapabilities;
	    keywords?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Tool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.icon = source["icon"];
	        this.version = source["version"];
	        this.capabilities = this.convertValues(source["capabilities"], ToolCapabilities);
	        this.keywords = source["keywords"];
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
	
	export class ToolError {
	    code: string;
	    message: string;
	    detail?: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.detail = source["detail"];
	    }
	}
	export class ToolInput {
	    toolId: string;
	    payload: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ToolInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolId = source["toolId"];
	        this.payload = source["payload"];
	    }
	}
	export class ToolOutput {
	    success: boolean;
	    data?: any;
	    error?: ToolError;
	
	    static createFrom(source: any = {}) {
	        return new ToolOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.data = source["data"];
	        this.error = this.convertValues(source["error"], ToolError);
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

}

