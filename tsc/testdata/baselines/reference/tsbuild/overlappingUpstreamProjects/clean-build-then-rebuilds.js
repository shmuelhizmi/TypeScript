currentDirectory::/home/src/workspaces/solution
useCaseSensitiveFileNames::true
Input::
//// [/home/src/workspaces/solution/node_modules/@ws/client] -> /home/src/workspaces/solution/packages/client *new*
//// [/home/src/workspaces/solution/node_modules/@ws/core] -> /home/src/workspaces/solution/packages/core *new*
//// [/home/src/workspaces/solution/package.json] *new* 
{
    "name": "solution",
    "private": true,
    "workspaces": ["packages/*"]
}
//// [/home/src/workspaces/solution/packages/app/package.json] *new* 
{
    "name": "@ws/app",
    "version": "1.0.0",
    "type": "module",
    "exports": {
        ".": {
            "types": "./dist/index.d.ts",
            "default": "./dist/index.js"
        }
    }
}
//// [/home/src/workspaces/solution/packages/app/src/index.ts] *new* 
import { greet } from "@ws/core";

export const message: string = greet("app");
//// [/home/src/workspaces/solution/packages/app/tsconfig.json] *new* 
{
    "compilerOptions": {
        "composite": true,
        "module": "ESNext",
        "moduleResolution": "Bundler",
        "target": "ES2022",
        "outDir": "./dist",
        "rootDir": "./src",
        "strict": true
    },
    "include": ["src/**/*"],
    "references": [{ "path": "../core" }, { "path": "../client" }]
}
//// [/home/src/workspaces/solution/packages/client/package.json] *new* 
{
    "name": "@ws/client",
    "version": "1.0.0",
    "type": "module",
    "exports": {
        ".": {
            "types": "./dist/index.d.ts",
            "default": "./dist/index.js"
        }
    }
}
//// [/home/src/workspaces/solution/packages/client/src/index.ts] *new* 
import { greet } from "@ws/core";

export const clientMessage: string = greet("client");
//// [/home/src/workspaces/solution/packages/client/tsconfig.json] *new* 
{
    "compilerOptions": {
        "composite": true,
        "module": "ESNext",
        "moduleResolution": "Bundler",
        "target": "ES2022",
        "outDir": "./dist",
        "rootDir": "./src",
        "strict": true
    },
    "include": ["src/**/*"],
    "references": [{ "path": "../core" }]
}
//// [/home/src/workspaces/solution/packages/core/package.json] *new* 
{
    "name": "@ws/core",
    "version": "1.0.0",
    "type": "module",
    "exports": {
        ".": {
            "types": "./dist/index.d.ts",
            "default": "./dist/index.js"
        }
    }
}
//// [/home/src/workspaces/solution/packages/core/src/index.ts] *new* 
export function greet(name: string): string {
    return "hello " + name;
}
//// [/home/src/workspaces/solution/packages/core/tsconfig.json] *new* 
{
    "compilerOptions": {
        "composite": true,
        "module": "ESNext",
        "moduleResolution": "Bundler",
        "target": "ES2022",
        "outDir": "./dist",
        "rootDir": "./src",
        "strict": true
    },
    "include": ["src/**/*"],
    "references": []
}
//// [/home/src/workspaces/solution/tsconfig.json] *new* 
{
    "files": [],
    "references": [
        { "path": "packages/core" },
        { "path": "packages/client" },
        { "path": "packages/app" }
    ]
}

tsgo --b --verbose
ExitStatus:: Success
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/core/tsconfig.json
    * packages/client/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/core/tsconfig.json' is out of date because output file 'packages/core/tsconfig.tsbuildinfo' does not exist

[[90mHH:MM:SS AM[0m] Building project 'packages/core/tsconfig.json'...

[[90mHH:MM:SS AM[0m] Project 'packages/client/tsconfig.json' is out of date because output file 'packages/client/tsconfig.tsbuildinfo' does not exist

[[90mHH:MM:SS AM[0m] Building project 'packages/client/tsconfig.json'...

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is out of date because output file 'packages/app/tsconfig.tsbuildinfo' does not exist

[[90mHH:MM:SS AM[0m] Building project 'packages/app/tsconfig.json'...

//// [/home/src/tslibs/TS/Lib/lib.es2022.full.d.ts] *Lib*
/// <reference no-default-lib="true"/>
interface Boolean {}
interface Function {}
interface CallableFunction {}
interface NewableFunction {}
interface IArguments {}
interface Number { toExponential: any; }
interface Object {}
interface RegExp {}
interface String { charAt: any; }
interface Array<T> { length: number; [n: number]: T; }
interface ReadonlyArray<T> {}
interface SymbolConstructor {
    (desc?: string | number): symbol;
    for(name: string): symbol;
    readonly toStringTag: symbol;
}
declare var Symbol: SymbolConstructor;
interface Symbol {
    readonly [Symbol.toStringTag]: string;
}
declare const console: { log(msg: any): void; };
//// [/home/src/workspaces/solution/packages/app/dist/index.d.ts] *new* 
export declare const message: string;

//// [/home/src/workspaces/solution/packages/app/dist/index.js] *new* 
import { greet } from "@ws/core";
export const message = greet("app");

//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo] *new* 
{"version":"FakeTSVersion","root":[3],"packageJsons":["./package.json","../core/package.json"],"fileNames":["lib.es2022.full.d.ts","../core/dist/index.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",{"version":"9e7ba452083270ec284c3a05000f6c7b-import { greet } from \"@ws/core\";\n\nexport const message: string = greet(\"app\");","signature":"62230bd755b106e316420041014346ca-export declare const message: string;\n","impliedNodeFormat":1}],"fileIdsList":[[2]],"options":{"composite":true,"module":99,"outDir":"./dist","rootDir":"./src","strict":true,"target":9},"referencedMap":[[3,1]],"latestChangedDtsFile":"./dist/index.d.ts"}
//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo.readable.baseline.txt] *new* 
{
  "version": "FakeTSVersion",
  "root": [
    {
      "files": [
        "./src/index.ts"
      ],
      "original": 3
    }
  ],
  "packageJsons": [
    "./package.json",
    "../core/package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "../core/dist/index.d.ts",
    "./src/index.ts"
  ],
  "fileInfos": [
    {
      "fileName": "lib.es2022.full.d.ts",
      "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "signature": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "affectsGlobalScope": true,
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
        "affectsGlobalScope": true,
        "impliedNodeFormat": 1
      }
    },
    {
      "fileName": "../core/dist/index.d.ts",
      "version": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS"
    },
    {
      "fileName": "./src/index.ts",
      "version": "9e7ba452083270ec284c3a05000f6c7b-import { greet } from \"@ws/core\";\n\nexport const message: string = greet(\"app\");",
      "signature": "62230bd755b106e316420041014346ca-export declare const message: string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "9e7ba452083270ec284c3a05000f6c7b-import { greet } from \"@ws/core\";\n\nexport const message: string = greet(\"app\");",
        "signature": "62230bd755b106e316420041014346ca-export declare const message: string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "fileIdsList": [
    [
      "../core/dist/index.d.ts"
    ]
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "referencedMap": {
    "./src/index.ts": [
      "../core/dist/index.d.ts"
    ]
  },
  "latestChangedDtsFile": "./dist/index.d.ts",
  "size": 1480
}
//// [/home/src/workspaces/solution/packages/client/dist/index.d.ts] *new* 
export declare const clientMessage: string;

//// [/home/src/workspaces/solution/packages/client/dist/index.js] *new* 
import { greet } from "@ws/core";
export const clientMessage = greet("client");

//// [/home/src/workspaces/solution/packages/client/tsconfig.tsbuildinfo] *new* 
{"version":"FakeTSVersion","root":[3],"packageJsons":["./package.json","../core/package.json"],"fileNames":["lib.es2022.full.d.ts","../core/dist/index.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",{"version":"7ad9a0976a81f178464d97ecca3c8cd3-import { greet } from \"@ws/core\";\n\nexport const clientMessage: string = greet(\"client\");","signature":"9ad25ec859c3d0a47e83d2cf08be21bb-export declare const clientMessage: string;\n","impliedNodeFormat":1}],"fileIdsList":[[2]],"options":{"composite":true,"module":99,"outDir":"./dist","rootDir":"./src","strict":true,"target":9},"referencedMap":[[3,1]],"latestChangedDtsFile":"./dist/index.d.ts"}
//// [/home/src/workspaces/solution/packages/client/tsconfig.tsbuildinfo.readable.baseline.txt] *new* 
{
  "version": "FakeTSVersion",
  "root": [
    {
      "files": [
        "./src/index.ts"
      ],
      "original": 3
    }
  ],
  "packageJsons": [
    "./package.json",
    "../core/package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "../core/dist/index.d.ts",
    "./src/index.ts"
  ],
  "fileInfos": [
    {
      "fileName": "lib.es2022.full.d.ts",
      "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "signature": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "affectsGlobalScope": true,
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
        "affectsGlobalScope": true,
        "impliedNodeFormat": 1
      }
    },
    {
      "fileName": "../core/dist/index.d.ts",
      "version": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS"
    },
    {
      "fileName": "./src/index.ts",
      "version": "7ad9a0976a81f178464d97ecca3c8cd3-import { greet } from \"@ws/core\";\n\nexport const clientMessage: string = greet(\"client\");",
      "signature": "9ad25ec859c3d0a47e83d2cf08be21bb-export declare const clientMessage: string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "7ad9a0976a81f178464d97ecca3c8cd3-import { greet } from \"@ws/core\";\n\nexport const clientMessage: string = greet(\"client\");",
        "signature": "9ad25ec859c3d0a47e83d2cf08be21bb-export declare const clientMessage: string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "fileIdsList": [
    [
      "../core/dist/index.d.ts"
    ]
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "referencedMap": {
    "./src/index.ts": [
      "../core/dist/index.d.ts"
    ]
  },
  "latestChangedDtsFile": "./dist/index.d.ts",
  "size": 1495
}
//// [/home/src/workspaces/solution/packages/core/dist/index.d.ts] *new* 
export declare function greet(name: string): string;

//// [/home/src/workspaces/solution/packages/core/dist/index.js] *new* 
export function greet(name) {
    return "hello " + name;
}

//// [/home/src/workspaces/solution/packages/core/tsconfig.tsbuildinfo] *new* 
{"version":"FakeTSVersion","root":[2],"packageJsons":["./package.json"],"fileNames":["lib.es2022.full.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},{"version":"0cb5926111ce2da24a0617196508db47-export function greet(name: string): string {\n    return \"hello \" + name;\n}","signature":"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n","impliedNodeFormat":1}],"options":{"composite":true,"module":99,"outDir":"./dist","rootDir":"./src","strict":true,"target":9},"latestChangedDtsFile":"./dist/index.d.ts"}
//// [/home/src/workspaces/solution/packages/core/tsconfig.tsbuildinfo.readable.baseline.txt] *new* 
{
  "version": "FakeTSVersion",
  "root": [
    {
      "files": [
        "./src/index.ts"
      ],
      "original": 2
    }
  ],
  "packageJsons": [
    "./package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "./src/index.ts"
  ],
  "fileInfos": [
    {
      "fileName": "lib.es2022.full.d.ts",
      "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "signature": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "affectsGlobalScope": true,
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
        "affectsGlobalScope": true,
        "impliedNodeFormat": 1
      }
    },
    {
      "fileName": "./src/index.ts",
      "version": "0cb5926111ce2da24a0617196508db47-export function greet(name: string): string {\n    return \"hello \" + name;\n}",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "0cb5926111ce2da24a0617196508db47-export function greet(name: string): string {\n    return \"hello \" + name;\n}",
        "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "latestChangedDtsFile": "./dist/index.d.ts",
  "size": 1306
}

packages/core/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/tslibs/TS/Lib/lib.es2022.full.d.ts
*refresh*    /home/src/workspaces/solution/packages/core/src/index.ts
Signatures::
(stored at emit) /home/src/workspaces/solution/packages/core/src/index.ts

packages/client/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/tslibs/TS/Lib/lib.es2022.full.d.ts
*refresh*    /home/src/workspaces/solution/packages/core/dist/index.d.ts
*refresh*    /home/src/workspaces/solution/packages/client/src/index.ts
Signatures::
(stored at emit) /home/src/workspaces/solution/packages/client/src/index.ts

packages/app/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/tslibs/TS/Lib/lib.es2022.full.d.ts
*refresh*    /home/src/workspaces/solution/packages/core/dist/index.d.ts
*refresh*    /home/src/workspaces/solution/packages/app/src/index.ts
Signatures::
(stored at emit) /home/src/workspaces/solution/packages/app/src/index.ts


Edit [0]:: no change

tsgo --b --verbose
ExitStatus:: Success
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/core/tsconfig.json
    * packages/client/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/core/tsconfig.json' is up to date because newest input 'packages/core/src/index.ts' is older than output 'packages/core/tsconfig.tsbuildinfo'

[[90mHH:MM:SS AM[0m] Project 'packages/client/tsconfig.json' is up to date because newest input 'packages/client/src/index.ts' is older than output 'packages/client/tsconfig.tsbuildinfo'

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is up to date because newest input 'packages/app/src/index.ts' is older than output 'packages/app/tsconfig.tsbuildinfo'




Edit [1]:: change core without changing its declarations
//// [/home/src/workspaces/solution/packages/core/src/index.ts] *modified* 
export function greet(name: string): string {
    return "hi " + name;
}

tsgo --b --verbose
ExitStatus:: Success
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/core/tsconfig.json
    * packages/client/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/core/tsconfig.json' is out of date because output 'packages/core/tsconfig.tsbuildinfo' is older than input 'packages/core/src/index.ts'

[[90mHH:MM:SS AM[0m] Building project 'packages/core/tsconfig.json'...

[[90mHH:MM:SS AM[0m] Project 'packages/client/tsconfig.json' is up to date with .d.ts files from its dependencies

[[90mHH:MM:SS AM[0m] Updating output timestamps of project 'packages/client/tsconfig.json'...

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is up to date with .d.ts files from its dependencies

[[90mHH:MM:SS AM[0m] Updating output timestamps of project 'packages/app/tsconfig.json'...

//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo] *mTime changed*
//// [/home/src/workspaces/solution/packages/client/tsconfig.tsbuildinfo] *mTime changed*
//// [/home/src/workspaces/solution/packages/core/dist/index.js] *modified* 
export function greet(name) {
    return "hi " + name;
}

//// [/home/src/workspaces/solution/packages/core/tsconfig.tsbuildinfo] *modified* 
{"version":"FakeTSVersion","root":[2],"packageJsons":["./package.json"],"fileNames":["lib.es2022.full.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},{"version":"cd0c8fec0d076792f613cf3f3de920f0-export function greet(name: string): string {\n    return \"hi \" + name;\n}","signature":"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n","impliedNodeFormat":1}],"options":{"composite":true,"module":99,"outDir":"./dist","rootDir":"./src","strict":true,"target":9},"latestChangedDtsFile":"./dist/index.d.ts"}
//// [/home/src/workspaces/solution/packages/core/tsconfig.tsbuildinfo.readable.baseline.txt] *modified* 
{
  "version": "FakeTSVersion",
  "root": [
    {
      "files": [
        "./src/index.ts"
      ],
      "original": 2
    }
  ],
  "packageJsons": [
    "./package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "./src/index.ts"
  ],
  "fileInfos": [
    {
      "fileName": "lib.es2022.full.d.ts",
      "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "signature": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "affectsGlobalScope": true,
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
        "affectsGlobalScope": true,
        "impliedNodeFormat": 1
      }
    },
    {
      "fileName": "./src/index.ts",
      "version": "cd0c8fec0d076792f613cf3f3de920f0-export function greet(name: string): string {\n    return \"hi \" + name;\n}",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "cd0c8fec0d076792f613cf3f3de920f0-export function greet(name: string): string {\n    return \"hi \" + name;\n}",
        "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "latestChangedDtsFile": "./dist/index.d.ts",
  "size": 1303
}

packages/core/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/workspaces/solution/packages/core/src/index.ts
Signatures::
(computed .d.ts) /home/src/workspaces/solution/packages/core/src/index.ts


Edit [2]:: introduce an error in app
//// [/home/src/workspaces/solution/packages/app/src/index.ts] *modified* 
import { greet } from "@ws/core";

export const message: number = greet("app");

tsgo --b --verbose
ExitStatus:: DiagnosticsPresent_OutputsGenerated
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/core/tsconfig.json
    * packages/client/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/core/tsconfig.json' is up to date because newest input 'packages/core/src/index.ts' is older than output 'packages/core/tsconfig.tsbuildinfo'

[[90mHH:MM:SS AM[0m] Project 'packages/client/tsconfig.json' is up to date because newest input 'packages/client/src/index.ts' is older than output 'packages/client/tsconfig.tsbuildinfo'

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is out of date because output 'packages/app/tsconfig.tsbuildinfo' is older than input 'packages/app/src/index.ts'

[[90mHH:MM:SS AM[0m] Building project 'packages/app/tsconfig.json'...

[96mpackages/app/src/index.ts[0m:[93m3[0m:[93m14[0m - [91merror[0m[90m TS2322: [0mType 'string' is not assignable to type 'number'.

[7m3[0m export const message: number = greet("app");
[7m [0m [91m             ~~~~~~~[0m


Found 1 error in packages/app/src/index.ts[90m:3[0m

//// [/home/src/workspaces/solution/packages/app/dist/index.d.ts] *modified* 
export declare const message: number;

//// [/home/src/workspaces/solution/packages/app/dist/index.js] *rewrite with same content*
//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo] *modified* 
{"version":"FakeTSVersion","root":[3],"packageJsons":["./package.json","../core/package.json"],"fileNames":["lib.es2022.full.d.ts","../core/dist/index.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",{"version":"2c168d27db470364ecf47227743cc6c8-import { greet } from \"@ws/core\";\n\nexport const message: number = greet(\"app\");","signature":"5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n","impliedNodeFormat":1}],"fileIdsList":[[2]],"options":{"composite":true,"module":99,"outDir":"./dist","rootDir":"./src","strict":true,"target":9},"referencedMap":[[3,1]],"semanticDiagnosticsPerFile":[[3,[{"pos":48,"end":55,"code":2322,"category":1,"messageKey":"Type_0_is_not_assignable_to_type_1_2322","messageArgs":["string","number"]}]]],"latestChangedDtsFile":"./dist/index.d.ts"}
//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo.readable.baseline.txt] *modified* 
{
  "version": "FakeTSVersion",
  "root": [
    {
      "files": [
        "./src/index.ts"
      ],
      "original": 3
    }
  ],
  "packageJsons": [
    "./package.json",
    "../core/package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "../core/dist/index.d.ts",
    "./src/index.ts"
  ],
  "fileInfos": [
    {
      "fileName": "lib.es2022.full.d.ts",
      "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "signature": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
      "affectsGlobalScope": true,
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };",
        "affectsGlobalScope": true,
        "impliedNodeFormat": 1
      }
    },
    {
      "fileName": "../core/dist/index.d.ts",
      "version": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS"
    },
    {
      "fileName": "./src/index.ts",
      "version": "2c168d27db470364ecf47227743cc6c8-import { greet } from \"@ws/core\";\n\nexport const message: number = greet(\"app\");",
      "signature": "5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "2c168d27db470364ecf47227743cc6c8-import { greet } from \"@ws/core\";\n\nexport const message: number = greet(\"app\");",
        "signature": "5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "fileIdsList": [
    [
      "../core/dist/index.d.ts"
    ]
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "referencedMap": {
    "./src/index.ts": [
      "../core/dist/index.d.ts"
    ]
  },
  "semanticDiagnosticsPerFile": [
    [
      "./src/index.ts",
      [
        {
          "pos": 48,
          "end": 55,
          "code": 2322,
          "category": 1,
          "messageKey": "Type_0_is_not_assignable_to_type_1_2322",
          "messageArgs": [
            "string",
            "number"
          ]
        }
      ]
    ]
  ],
  "latestChangedDtsFile": "./dist/index.d.ts",
  "size": 1651
}

packages/app/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/workspaces/solution/packages/app/src/index.ts
Signatures::
(computed .d.ts) /home/src/workspaces/solution/packages/app/src/index.ts
