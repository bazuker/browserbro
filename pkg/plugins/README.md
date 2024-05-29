# Plugins ⚙️

## Overview

A plugin is a Golang package with a struct that implements [the Plugin interface](plugins.go).

#### Name
All plugins must have URL-safe names that will be used as handlers during requests routing.
For example a plugin named `foobar` will be accessible at `/api/v1/plugins/foobar`.


#### Input params
Plugins are intentionally isolated from the router and should not have access 
to an incoming request context or a response writer.
When a new request is received, the request body is processed by the manager and all necessary inputs will be conveniently available
to plugins in `params map[string]any`.

#### Output params
As a plugin completes its execution, the results of the execution should be written to the `output map[string]any`.
It then will be encoded and written to the response body by the manager without alteration.
There is no need to include plugin's name in the output as the manager will automatically include it by default:
```
{
  "plugin name": {
   ... output will be put here
  }
}
```

#### Errors handling
If during the execution a plugin encounters an error, the output will be discarded and only the error message will be returned along with the 500 Internal Error status.

#### Files
If a plugin should provide files as its output, a plugin should generate unique file names and save them at the `BasePath`
available in `FileStore`. The file names should be then written to the output. 
The end user will be able to access the files using the Files API. [See README - Files section](..%2F..%2FREADME.md).

It is recommended to generate URL-safe file names.

The`FileStore` usage example can found in the [screenshot](screenshot%2Fscreenshot.go) plugin.

## Creating a plugin
1. Create a new package under `plugins` directory with a unique name for your plugin
2. Create a plugin struct that implements [the Plugin interface](plugins.go). 
3. Register your plugin in `initPlugins` function in [main.go](..%2F..%2Fmain.go)
4. Add README.md file to your plugin's directory. 
Make sure to include name, required input parameters and output format.
