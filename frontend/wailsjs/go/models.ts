export namespace domain {
	
	export class CommitWithFiles {
	    hash: string;
	    subject: string;
	    author: string;
	    date: string;
	    files: string[];
	    isMerge: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CommitWithFiles(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hash = source["hash"];
	        this.subject = source["subject"];
	        this.author = source["author"];
	        this.date = source["date"];
	        this.files = source["files"];
	        this.isMerge = source["isMerge"];
	    }
	}
	export class FileNode {
	    name: string;
	    path: string;
	    relPath: string;
	    isDir: boolean;
	    size: number;
	    children?: FileNode[];
	    isGitignored: boolean;
	    isCustomIgnored: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.relPath = source["relPath"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	        this.children = this.convertValues(source["children"], FileNode);
	        this.isGitignored = source["isGitignored"];
	        this.isCustomIgnored = source["isCustomIgnored"];
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
	export class FileStatus {
	    path: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new FileStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.status = source["status"];
	    }
	}
	export class SettingsDTO {
	    customIgnoreRules: string;
	    customPromptRules: string;
	    openAIAPIKey: string;
	    geminiAPIKey: string;
	    openRouterAPIKey: string;
	    localAIAPIKey: string;
	    localAIHost: string;
	    localAIModelName: string;
	    selectedProvider: string;
	    selectedModels: Record<string, string>;
	    availableModels: Record<string, string[]>;
	    useGitignore: boolean;
	    useCustomIgnore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.customIgnoreRules = source["customIgnoreRules"];
	        this.customPromptRules = source["customPromptRules"];
	        this.openAIAPIKey = source["openAIAPIKey"];
	        this.geminiAPIKey = source["geminiAPIKey"];
	        this.openRouterAPIKey = source["openRouterAPIKey"];
	        this.localAIAPIKey = source["localAIAPIKey"];
	        this.localAIHost = source["localAIHost"];
	        this.localAIModelName = source["localAIModelName"];
	        this.selectedProvider = source["selectedProvider"];
	        this.selectedModels = source["selectedModels"];
	        this.availableModels = source["availableModels"];
	        this.useGitignore = source["useGitignore"];
	        this.useCustomIgnore = source["useCustomIgnore"];
	    }
	}

}

