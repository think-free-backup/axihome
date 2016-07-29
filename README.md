

### Axihome 3

-----------------------------------------------------------------------------------------------------------


Axihome is a home automation / iot solution written entirely in Go and developped initialy for my personal use.
It can run on any linux platform (windows not tested) but the main architecture is a raspberry pi (2 or +)
I have it drinving the lights, shutters, pool (motor, temperature, uv index, lights), heating, fitbit integration, custom smartmirror, custom smartalarmclock, room temperature, humidity, ...


Introduction
-----------------------------------------------------

The server is divided in various process and comunicate with a custom json protocol over tcp or websocket.   
It is divided in three parts :

- The core server (TCP : 3330, WS : 3331) : for core components like notifications managment, smartheating, cronjobs, acond, ... Every process in the core receive the variable change message.
- The backend server (TCP : 3332, WS : 3333) : clients of the backend server are responsible of gatering data of differents sources like the ipx800, x10modules, waze, yahoo weather, ...
- The frontend server (TCP : 3334, WS : 3335) : graphical clients connect to this server. Every process in the frontend receive the variable change message and the client configuration message.


Quick start
-----------------------------------------------------

- Download release .run and execute it on a raspbian image
- Create a Instances.json and a Variables.json file
- go to /etc/axihome
- Run ./config/importer.sh path to your config folder

Minimal config files
-----------------------------------------------------

Config.json

    {
        "buckets" : []
    }

Instances.json

    {
        "time" : {
            "name" : "time",
            "backend" : "time",
            "params" : {},
            "run" : true
        },    
        "historic" : {
            "name" : "historic",
            "backend" : "historic",
            "params" : {},
            "run" : true
        },
        "chart" : {
            "name" : "chart",
            "backend" : "chart",
            "params" : {},
            "run" : true
        }
    }

Variables.json

    {
        "server.time" : { "addr" : "", "shortname" : "", "name" : "", "backend": "time", "type" : "float64", "default" : null, "analog" : false },
    }
    
Processes managment
-----------------------------------------------------

The main process is responsible of managing all the others processes.   
You can connect to the manager at http://serverip:3340

Core and backend processes are defined in Instances.json


Variables definition
-----------------------------------------------------

    {
        "variablename" : { "addr" : "", "shortname" : "", "name" : "", "backend": "backendname", "type" : "string|float64|bool", "default" : null, "analog" : false }
    }

- The variable name is the name used in the frontend and the core
- The addr is the name used in the backend (usefull if you want to rename a backend variable for your frontend). If blank it use the variablename instead (faster).
- Backend is the process name of the backend 
- Type : you only have 3 types availables : string, float64 and bool
- Default : default value on clean startup (if you have no rtdb)
- Analog : used for historic database : false -> write on change, true -> write every 5 minutes


Developping core or backend processes
-----------------------------------------------------

- go get github.com/think-free/axihome
- run build.sh from your GOPATH
- run package.sh from your GOPATH if you have makeself installed

The fastest way to devellop a core/backend process is to lock at an existing one, they are small, self contained.   
You have to know some rules :

__Backend processes call variables.set to write a variable__

    jsontools.GenerateRpcMessage(&sendChannel, "variables", "set", {"key": "value"}, "mydomain", "axihome")

__To write to a backend call variables.write__

    jsontools.GenerateRpcMessage(&sendChannel, "variables", "write", {"key": "value"}, "mydomain", "axihome")

__To get a variable from a bucket (answered by a setVar call to the callee)__

    jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getVar", {"bucket" : "name", "variable" : "name"}, "mydomain", "axihome")

__To get all the content of a bucket__

    jsontools.GenerateRpcMessage(&sendChannel, "bucket", "getAll", "BucketName", "mydomain", "axihome")

__To set a value of a variable of a bucket__

    jsontools.GenerateRpcMessage(&sendChannel, "bucket", "setVar", {"bucket" : "BucketName", "variable" : "name", "value" : "value"}, "mydomain", "axihome")

__To set the content of a bucket (bucket content will be erased)__

    jsontools.GenerateRpcMessage(&sendChannel, "bucket", "setAll", {"bucket" : "BucketName", "content" : {"key" : "value"}}, "mydomain", "axihome")

Roadmap
-----------------------------------------------------

- Separation of the main application from the core process was great at the beginning of the development for stability but as the application grow and more you have variables changing, it become a bottle neck. One big change would be to integrate all the funcionalities of the core processes in the main executable.
- Another big change would be the implementation of variables subscriptions, as for now every clients (and core processes) are receiving every changes
- A web interface for easy configuration would be a big win.
- Documentation
- Documentation
- Documentation
- Documentation
- Documentation
- ...
