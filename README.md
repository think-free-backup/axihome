
------------------------------------------------------------------------------

# Axihome 3

------------------------------------------------------------------------------


The server is divided in various process and comunicate with a custom json protocol over tcp or websocket

Main process
-----------------------------------------------------

The main process of the application create 3 diferents json servers :

- The core server : for core components like notifications managment, smartheating, cronjobs, acond, ... Every process in the core receive the variable change message
- The backend server : clients of the backend server are responsible of gatering data of differents sources like the ipx800, x10modules, waze, yahoo weather, ...
- The frontend server : graphical clients connect to this server. Every process in the frontend receive the variable change message and the client configuration message

The main process is also responsible of the life of all the others processes you can connect to http://serverip:3340 to see the running process, to start/stop/restart them. To have the state in json format you can request http://serverip:3340/getstate and to start/stop/restart process with direct request to : http://serverip:3340/start?process=name&instance=name

The configuration is defined like this (Instances.json) :

    {
        "notification" : {
            "name" : "notification",
            "backend" : "notification",
            "params" : {},
            "run" : true
        }
    }

Currently is also manage all the databases (configuration, variables, historic) but this part will be moved to a core process in the future

Data core process (for now it's the main process)
-----------------------------------------------------

The variable list is defined like this (Variables.json) : 

{
    "variablename" : { "shortname" : "", "name" : "", "backend": "backendname", "type" : "float64", "analog" : true }
}

With variablename = the name of the variable coming from the backend process, and backendname the instance name of the backend.
Analog variables are saved in historic every 5 minutes digital are saved on change and a full rtdb dump is saved every hours.

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



Notification core process
-----------------------------------------------------

This process is responsible of sending notification to subscribed devices. For now it support telegram and a custom tts system (called say) accessible by REST calls.

__NotificationGeneral.json__

This is the general configuration for notifications

    {
        "telegram" : "telegram bot key",
        "say" : null
    }

__NotificationMessages.json__

This is the messages definition for notifications

    {
        "messagename" : {"shortText" : "", "mediumText" : "", "largeText" : "", "sound" : "", "image" : "", "url" : ""}
    }

__NotificationDevices.json__

All the devices ables to receive notifications

    {
        "devicename" : {"notifier": "say", "desc" : "First say device", "url" : "192.168.1.2"}
    }

__NotificationMessagesSubscriptions.json__

Subscriptions for messages

    {
        "messagename" : [ {"dev": "devicename", "type" : "mediumText", "active" : "booleanvariablename"} ]
    }

To send a notification use the call : jsontools.GenerateRpcMessage(sendChannel, "notifier", "send", "messagename", "mydomain", "notification.core.axihome")

Variables notifications core process
-----------------------------------------------------

This process send a notification if the variablename's value respect the condition. Conditions are : >, >=, =, <=, <, !=

    {
        "variablename" : [{"cond" : "!=", "val" : "conditionvalue", "notif" : "messagename"}]
    }
