currentDirectory::/home/src/workspaces/solution
useCaseSensitiveFileNames::true
Input::
//// [/home/src/workspaces/solution/node_modules/lib] -> /home/src/workspaces/solution/packages/lib *new*
//// [/home/src/workspaces/solution/package.json] *new* 
{
    "name": "solution",
    "private": true,
    "workspaces": ["packages/*"]
}
//// [/home/src/workspaces/solution/packages/app/package.json] *new* 
{
    "name": "app",
    "version": "1.0.0",
    "type": "module",
    "dependencies": {
        "lib": "workspace:*"
    }
}
//// [/home/src/workspaces/solution/packages/app/src/index.ts] *new* 
import { greet } from "lib";

export const message: string = greet("world");
//// [/home/src/workspaces/solution/packages/app/tsconfig.json] *new* 
{
    "compilerOptions": {
        "composite": true,
        "module": "ESNext",
        "moduleResolution": "Bundler",
        "target": "ES2022",
        "outDir": "./out",
        "rootDir": "./src",
        "strict": true
    },
    "include": ["src/**/*"]
}
//// [/home/src/workspaces/solution/packages/lib/package.json] *new* 
{
    "name": "lib",
    "version": "1.0.0",
    "type": "module",
    "exports": {
        ".": "./src/index.ts"
    }
}
//// [/home/src/workspaces/solution/packages/lib/src/index.ts] *new* 
export function greet(name: string): string {
    const count: number = name;
    return name;
}
//// [/home/src/workspaces/solution/packages/lib/tsconfig.json] *new* 
{
    "compilerOptions": {
        "composite": true,
        "module": "ESNext",
        "moduleResolution": "Bundler",
        "target": "ES2022",
        "outDir": "./out",
        "rootDir": "./src",
        "strict": true
    },
    "include": ["src/**/*"]
}
//// [/home/src/workspaces/solution/tsconfig.json] *new* 
{
    "files": [],
    "references": [
        { "path": "packages/lib" },
        { "path": "packages/app" }
    ]
}

tsgo --b --verbose
ExitStatus:: DiagnosticsPresent_OutputsGenerated
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/lib/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/lib/tsconfig.json' is out of date because output file 'packages/lib/tsconfig.tsbuildinfo' does not exist

[[90mHH:MM:SS AM[0m] Building project 'packages/lib/tsconfig.json'...

[96mpackages/lib/src/index.ts[0m:[93m2[0m:[93m11[0m - [91merror[0m[90m TS2322: [0mType 'string' is not assignable to type 'number'.

[7m2[0m     const count: number = name;
[7m [0m [91m          ~~~~~[0m

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is out of date because output file 'packages/app/tsconfig.tsbuildinfo' does not exist

[[90mHH:MM:SS AM[0m] Building project 'packages/app/tsconfig.json'...


Found 1 error in packages/lib/src/index.ts[90m:2[0m

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
//// [/home/src/workspaces/solution/packages/app/out/index.d.ts] *new* 
export declare const message: string;

//// [/home/src/workspaces/solution/packages/app/out/index.js] *new* 
import { greet } from "lib";
export const message = greet("world");

//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo] *new* 
{"version":"FakeTSVersion","root":[3],"packageJsons":["./package.json","../lib/package.json"],"fileNames":["lib.es2022.full.d.ts","../lib/src/index.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},"0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",{"version":"b4c2e8839b08ea8a80a84973aaa4363d-import { greet } from \"lib\";\n\nexport const message: string = greet(\"world\");","signature":"62230bd755b106e316420041014346ca-export declare const message: string;\n","impliedNodeFormat":1}],"fileIdsList":[[2]],"options":{"composite":true,"module":99,"outDir":"./out","rootDir":"./src","strict":true,"target":9},"referencedMap":[[3,1]],"latestChangedDtsFile":"./out/index.d.ts"}
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
    "../lib/package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "../lib/src/index.ts",
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
      "fileName": "../lib/src/index.ts",
      "version": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
      "signature": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
      "impliedNodeFormat": "CommonJS"
    },
    {
      "fileName": "./src/index.ts",
      "version": "b4c2e8839b08ea8a80a84973aaa4363d-import { greet } from \"lib\";\n\nexport const message: string = greet(\"world\");",
      "signature": "62230bd755b106e316420041014346ca-export declare const message: string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "b4c2e8839b08ea8a80a84973aaa4363d-import { greet } from \"lib\";\n\nexport const message: string = greet(\"world\");",
        "signature": "62230bd755b106e316420041014346ca-export declare const message: string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "fileIdsList": [
    [
      "../lib/src/index.ts"
    ]
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./out",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "referencedMap": {
    "./src/index.ts": [
      "../lib/src/index.ts"
    ]
  },
  "latestChangedDtsFile": "./out/index.d.ts",
  "size": 1515
}
//// [/home/src/workspaces/solution/packages/lib/out/index.d.ts] *new* 
export declare function greet(name: string): string;

//// [/home/src/workspaces/solution/packages/lib/out/index.js] *new* 
export function greet(name) {
    const count = name;
    return name;
}

//// [/home/src/workspaces/solution/packages/lib/tsconfig.tsbuildinfo] *new* 
{"version":"FakeTSVersion","root":[2],"packageJsons":["./package.json"],"fileNames":["lib.es2022.full.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},{"version":"0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}","signature":"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n","impliedNodeFormat":1}],"options":{"composite":true,"module":99,"outDir":"./out","rootDir":"./src","strict":true,"target":9},"semanticDiagnosticsPerFile":[[2,[{"pos":56,"end":61,"code":2322,"category":1,"messageKey":"Type_0_is_not_assignable_to_type_1_2322","messageArgs":["string","number"]}]]],"latestChangedDtsFile":"./out/index.d.ts"}
//// [/home/src/workspaces/solution/packages/lib/tsconfig.tsbuildinfo.readable.baseline.txt] *new* 
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
      "version": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
        "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./out",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "semanticDiagnosticsPerFile": [
    [
      "./src/index.ts",
      [
        {
          "pos": 56,
          "end": 61,
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
  "latestChangedDtsFile": "./out/index.d.ts",
  "size": 1495
}

packages/lib/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/tslibs/TS/Lib/lib.es2022.full.d.ts
*refresh*    /home/src/workspaces/solution/packages/lib/src/index.ts
Signatures::
(stored at emit) /home/src/workspaces/solution/packages/lib/src/index.ts

packages/app/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/tslibs/TS/Lib/lib.es2022.full.d.ts
*refresh*    /home/src/workspaces/solution/packages/lib/src/index.ts
*refresh*    /home/src/workspaces/solution/packages/app/src/index.ts
Signatures::
(stored at emit) /home/src/workspaces/solution/packages/app/src/index.ts


Edit [0]:: no change

tsgo --b --verbose
ExitStatus:: DiagnosticsPresent_OutputsGenerated
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/lib/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/lib/tsconfig.json' is out of date because buildinfo file 'packages/lib/tsconfig.tsbuildinfo' indicates that program needs to report errors.

[[90mHH:MM:SS AM[0m] Building project 'packages/lib/tsconfig.json'...

[96mpackages/lib/src/index.ts[0m:[93m2[0m:[93m11[0m - [91merror[0m[90m TS2322: [0mType 'string' is not assignable to type 'number'.

[7m2[0m     const count: number = name;
[7m [0m [91m          ~~~~~[0m

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is up to date because newest input 'packages/app/src/index.ts' is older than output 'packages/app/tsconfig.tsbuildinfo'


Found 1 error in packages/lib/src/index.ts[90m:2[0m


packages/lib/tsconfig.json::
SemanticDiagnostics::
Signatures::


Edit [1]:: fix the error in lib
//// [/home/src/workspaces/solution/packages/lib/src/index.ts] *modified* 
export function greet(name: string): string {
    const count: string = name;
    return name;
}

tsgo --b --verbose
ExitStatus:: Success
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/lib/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/lib/tsconfig.json' is out of date because buildinfo file 'packages/lib/tsconfig.tsbuildinfo' indicates that program needs to report errors.

[[90mHH:MM:SS AM[0m] Building project 'packages/lib/tsconfig.json'...

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is out of date because output 'packages/app/tsconfig.tsbuildinfo' is older than input 'packages/lib/src/index.ts'

[[90mHH:MM:SS AM[0m] Building project 'packages/app/tsconfig.json'...

//// [/home/src/workspaces/solution/packages/app/out/index.js] *rewrite with same content*
//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo] *modified* 
{"version":"FakeTSVersion","root":[3],"packageJsons":["./package.json","../lib/package.json"],"fileNames":["lib.es2022.full.d.ts","../lib/src/index.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},"52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",{"version":"b4c2e8839b08ea8a80a84973aaa4363d-import { greet } from \"lib\";\n\nexport const message: string = greet(\"world\");","signature":"62230bd755b106e316420041014346ca-export declare const message: string;\n","impliedNodeFormat":1}],"fileIdsList":[[2]],"options":{"composite":true,"module":99,"outDir":"./out","rootDir":"./src","strict":true,"target":9},"referencedMap":[[3,1]],"latestChangedDtsFile":"./out/index.d.ts"}
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
    "../lib/package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "../lib/src/index.ts",
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
      "fileName": "../lib/src/index.ts",
      "version": "52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",
      "signature": "52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",
      "impliedNodeFormat": "CommonJS"
    },
    {
      "fileName": "./src/index.ts",
      "version": "b4c2e8839b08ea8a80a84973aaa4363d-import { greet } from \"lib\";\n\nexport const message: string = greet(\"world\");",
      "signature": "62230bd755b106e316420041014346ca-export declare const message: string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "b4c2e8839b08ea8a80a84973aaa4363d-import { greet } from \"lib\";\n\nexport const message: string = greet(\"world\");",
        "signature": "62230bd755b106e316420041014346ca-export declare const message: string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "fileIdsList": [
    [
      "../lib/src/index.ts"
    ]
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./out",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "referencedMap": {
    "./src/index.ts": [
      "../lib/src/index.ts"
    ]
  },
  "latestChangedDtsFile": "./out/index.d.ts",
  "size": 1515
}
//// [/home/src/workspaces/solution/packages/lib/out/index.js] *rewrite with same content*
//// [/home/src/workspaces/solution/packages/lib/tsconfig.tsbuildinfo] *modified* 
{"version":"FakeTSVersion","root":[2],"packageJsons":["./package.json"],"fileNames":["lib.es2022.full.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},{"version":"52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}","signature":"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n","impliedNodeFormat":1}],"options":{"composite":true,"module":99,"outDir":"./out","rootDir":"./src","strict":true,"target":9},"latestChangedDtsFile":"./out/index.d.ts"}
//// [/home/src/workspaces/solution/packages/lib/tsconfig.tsbuildinfo.readable.baseline.txt] *modified* 
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
      "version": "52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",
        "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./out",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "latestChangedDtsFile": "./out/index.d.ts",
  "size": 1324
}

packages/lib/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/workspaces/solution/packages/lib/src/index.ts
Signatures::
(computed .d.ts) /home/src/workspaces/solution/packages/lib/src/index.ts

packages/app/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/workspaces/solution/packages/lib/src/index.ts
*refresh*    /home/src/workspaces/solution/packages/app/src/index.ts
Signatures::
(used version)   /home/src/workspaces/solution/packages/lib/src/index.ts
(computed .d.ts) /home/src/workspaces/solution/packages/app/src/index.ts


Edit [2]:: introduce an error in app
//// [/home/src/workspaces/solution/packages/app/src/index.ts] *modified* 
import { greet } from "lib";

export const message: number = greet("world");

tsgo --b --verbose
ExitStatus:: DiagnosticsPresent_OutputsGenerated
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/lib/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/lib/tsconfig.json' is up to date because newest input 'packages/lib/src/index.ts' is older than output 'packages/lib/tsconfig.tsbuildinfo'

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is out of date because output 'packages/app/tsconfig.tsbuildinfo' is older than input 'packages/app/src/index.ts'

[[90mHH:MM:SS AM[0m] Building project 'packages/app/tsconfig.json'...

[96mpackages/app/src/index.ts[0m:[93m3[0m:[93m14[0m - [91merror[0m[90m TS2322: [0mType 'string' is not assignable to type 'number'.

[7m3[0m export const message: number = greet("world");
[7m [0m [91m             ~~~~~~~[0m


Found 1 error in packages/app/src/index.ts[90m:3[0m

//// [/home/src/workspaces/solution/packages/app/out/index.d.ts] *modified* 
export declare const message: number;

//// [/home/src/workspaces/solution/packages/app/out/index.js] *rewrite with same content*
//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo] *modified* 
{"version":"FakeTSVersion","root":[3],"packageJsons":["./package.json","../lib/package.json"],"fileNames":["lib.es2022.full.d.ts","../lib/src/index.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},"52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",{"version":"c897ee5b2a4045dfc424b7da2e522af7-import { greet } from \"lib\";\n\nexport const message: number = greet(\"world\");","signature":"5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n","impliedNodeFormat":1}],"fileIdsList":[[2]],"options":{"composite":true,"module":99,"outDir":"./out","rootDir":"./src","strict":true,"target":9},"referencedMap":[[3,1]],"semanticDiagnosticsPerFile":[[3,[{"pos":43,"end":50,"code":2322,"category":1,"messageKey":"Type_0_is_not_assignable_to_type_1_2322","messageArgs":["string","number"]}]]],"latestChangedDtsFile":"./out/index.d.ts"}
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
    "../lib/package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "../lib/src/index.ts",
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
      "fileName": "../lib/src/index.ts",
      "version": "52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",
      "signature": "52fbb3b4ce882440ea091ca7d73a3a35-export function greet(name: string): string {\n    const count: string = name;\n    return name;\n}",
      "impliedNodeFormat": "CommonJS"
    },
    {
      "fileName": "./src/index.ts",
      "version": "c897ee5b2a4045dfc424b7da2e522af7-import { greet } from \"lib\";\n\nexport const message: number = greet(\"world\");",
      "signature": "5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "c897ee5b2a4045dfc424b7da2e522af7-import { greet } from \"lib\";\n\nexport const message: number = greet(\"world\");",
        "signature": "5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "fileIdsList": [
    [
      "../lib/src/index.ts"
    ]
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./out",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "referencedMap": {
    "./src/index.ts": [
      "../lib/src/index.ts"
    ]
  },
  "semanticDiagnosticsPerFile": [
    [
      "./src/index.ts",
      [
        {
          "pos": 43,
          "end": 50,
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
  "latestChangedDtsFile": "./out/index.d.ts",
  "size": 1686
}

packages/app/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/workspaces/solution/packages/app/src/index.ts
Signatures::
(computed .d.ts) /home/src/workspaces/solution/packages/app/src/index.ts


Edit [3]:: reintroduce the error in lib
//// [/home/src/workspaces/solution/packages/lib/src/index.ts] *modified* 
export function greet(name: string): string {
    const count: number = name;
    return name;
}

tsgo --b --verbose
ExitStatus:: DiagnosticsPresent_OutputsGenerated
Output::
[[90mHH:MM:SS AM[0m] Projects in this build: 
    * packages/lib/tsconfig.json
    * packages/app/tsconfig.json
    * tsconfig.json

[[90mHH:MM:SS AM[0m] Project 'packages/lib/tsconfig.json' is out of date because output 'packages/lib/tsconfig.tsbuildinfo' is older than input 'packages/lib/src/index.ts'

[[90mHH:MM:SS AM[0m] Building project 'packages/lib/tsconfig.json'...

[96mpackages/lib/src/index.ts[0m:[93m2[0m:[93m11[0m - [91merror[0m[90m TS2322: [0mType 'string' is not assignable to type 'number'.

[7m2[0m     const count: number = name;
[7m [0m [91m          ~~~~~[0m

[[90mHH:MM:SS AM[0m] Project 'packages/app/tsconfig.json' is out of date because buildinfo file 'packages/app/tsconfig.tsbuildinfo' indicates that program needs to report errors.

[[90mHH:MM:SS AM[0m] Building project 'packages/app/tsconfig.json'...

[96mpackages/app/src/index.ts[0m:[93m3[0m:[93m14[0m - [91merror[0m[90m TS2322: [0mType 'string' is not assignable to type 'number'.

[7m3[0m export const message: number = greet("world");
[7m [0m [91m             ~~~~~~~[0m


Found 2 errors in 2 files.

Errors  Files
     1  packages/app/src/index.ts[90m:3[0m
     1  packages/lib/src/index.ts[90m:2[0m

//// [/home/src/workspaces/solution/packages/app/out/index.js] *rewrite with same content*
//// [/home/src/workspaces/solution/packages/app/tsconfig.tsbuildinfo] *modified* 
{"version":"FakeTSVersion","root":[3],"packageJsons":["./package.json","../lib/package.json"],"fileNames":["lib.es2022.full.d.ts","../lib/src/index.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},"0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",{"version":"c897ee5b2a4045dfc424b7da2e522af7-import { greet } from \"lib\";\n\nexport const message: number = greet(\"world\");","signature":"5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n","impliedNodeFormat":1}],"fileIdsList":[[2]],"options":{"composite":true,"module":99,"outDir":"./out","rootDir":"./src","strict":true,"target":9},"referencedMap":[[3,1]],"semanticDiagnosticsPerFile":[[3,[{"pos":43,"end":50,"code":2322,"category":1,"messageKey":"Type_0_is_not_assignable_to_type_1_2322","messageArgs":["string","number"]}]]],"latestChangedDtsFile":"./out/index.d.ts"}
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
    "../lib/package.json"
  ],
  "fileNames": [
    "lib.es2022.full.d.ts",
    "../lib/src/index.ts",
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
      "fileName": "../lib/src/index.ts",
      "version": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
      "signature": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
      "impliedNodeFormat": "CommonJS"
    },
    {
      "fileName": "./src/index.ts",
      "version": "c897ee5b2a4045dfc424b7da2e522af7-import { greet } from \"lib\";\n\nexport const message: number = greet(\"world\");",
      "signature": "5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "c897ee5b2a4045dfc424b7da2e522af7-import { greet } from \"lib\";\n\nexport const message: number = greet(\"world\");",
        "signature": "5499e90f4fd064f99c703f815deed8d3-export declare const message: number;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "fileIdsList": [
    [
      "../lib/src/index.ts"
    ]
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./out",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "referencedMap": {
    "./src/index.ts": [
      "../lib/src/index.ts"
    ]
  },
  "semanticDiagnosticsPerFile": [
    [
      "./src/index.ts",
      [
        {
          "pos": 43,
          "end": 50,
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
  "latestChangedDtsFile": "./out/index.d.ts",
  "size": 1686
}
//// [/home/src/workspaces/solution/packages/lib/out/index.js] *rewrite with same content*
//// [/home/src/workspaces/solution/packages/lib/tsconfig.tsbuildinfo] *modified* 
{"version":"FakeTSVersion","root":[2],"packageJsons":["./package.json"],"fileNames":["lib.es2022.full.d.ts","./src/index.ts"],"fileInfos":[{"version":"8859c12c614ce56ba9a18e58384a198f-/// <reference no-default-lib=\"true\"/>\ninterface Boolean {}\ninterface Function {}\ninterface CallableFunction {}\ninterface NewableFunction {}\ninterface IArguments {}\ninterface Number { toExponential: any; }\ninterface Object {}\ninterface RegExp {}\ninterface String { charAt: any; }\ninterface Array<T> { length: number; [n: number]: T; }\ninterface ReadonlyArray<T> {}\ninterface SymbolConstructor {\n    (desc?: string | number): symbol;\n    for(name: string): symbol;\n    readonly toStringTag: symbol;\n}\ndeclare var Symbol: SymbolConstructor;\ninterface Symbol {\n    readonly [Symbol.toStringTag]: string;\n}\ndeclare const console: { log(msg: any): void; };","affectsGlobalScope":true,"impliedNodeFormat":1},{"version":"0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}","signature":"bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n","impliedNodeFormat":1}],"options":{"composite":true,"module":99,"outDir":"./out","rootDir":"./src","strict":true,"target":9},"semanticDiagnosticsPerFile":[[2,[{"pos":56,"end":61,"code":2322,"category":1,"messageKey":"Type_0_is_not_assignable_to_type_1_2322","messageArgs":["string","number"]}]]],"latestChangedDtsFile":"./out/index.d.ts"}
//// [/home/src/workspaces/solution/packages/lib/tsconfig.tsbuildinfo.readable.baseline.txt] *modified* 
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
      "version": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
      "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
      "impliedNodeFormat": "CommonJS",
      "original": {
        "version": "0b643ec087b7d089e2df4fe2c92857c5-export function greet(name: string): string {\n    const count: number = name;\n    return name;\n}",
        "signature": "bac8ffa27742700972667af00775dd44-export declare function greet(name: string): string;\n",
        "impliedNodeFormat": 1
      }
    }
  ],
  "options": {
    "composite": true,
    "module": 99,
    "outDir": "./out",
    "rootDir": "./src",
    "strict": true,
    "target": 9
  },
  "semanticDiagnosticsPerFile": [
    [
      "./src/index.ts",
      [
        {
          "pos": 56,
          "end": 61,
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
  "latestChangedDtsFile": "./out/index.d.ts",
  "size": 1495
}

packages/lib/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/workspaces/solution/packages/lib/src/index.ts
Signatures::
(computed .d.ts) /home/src/workspaces/solution/packages/lib/src/index.ts

packages/app/tsconfig.json::
SemanticDiagnostics::
*refresh*    /home/src/workspaces/solution/packages/lib/src/index.ts
*refresh*    /home/src/workspaces/solution/packages/app/src/index.ts
Signatures::
(used version)   /home/src/workspaces/solution/packages/lib/src/index.ts
(computed .d.ts) /home/src/workspaces/solution/packages/app/src/index.ts
