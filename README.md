# Note to the authors

## Chapter 2

### Product Info

- The `go_package` option must be specified in the `.proto` file. And in the command line, the `protoc-gen-go` plugin is deprecated, and the `--go-grpc_out` flag must be declared.
- The `UnimplementedProductInfoServer` must be embedded in the server struct definition.
- Add notes to explain what the `go_package` option is, why we need it and that we need to first push to the repository first in order to import the pb defintions
- `com.google.protobuf-gradle-plugin:0.8.10:` cannot be found as the plugin in the `build.gradle` file.
- protoc does not have support for Apple Silicon chips yet, error ` Could not find protoc-3.15.6-osx-aarch_64.exe` occurs 
