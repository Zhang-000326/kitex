package tpl

// ConfigTpl is the template for generating yaml config file.
var ConfigTpl = `Address: ":8888"
EnableDebugServer: true
DebugServerPort: "18888"
Log:
  Dir: log
  Loggers:
    - Name: default
      Level: info # Notice: change it to debug if needed in local development
      Outputs:
        - File
        - Agent
        # - Console # Notice: change it to debug if needed in local development, don't use this in production!
    - Name: rpcAccess
      Level: trace    # Notice: Not recommended for modification, otherwise may affect construction of call chain (tracing)
      Outputs:
        - File
        - Agent
    - Name: rpcCall
      Level: trace    # Notice: Not recommended for modification, otherwise may affect construction of call chain (tracing)
      Outputs:
        - File
        - Agent`
