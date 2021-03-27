# Note to the authors

## Chapter 2

### Product Info

- The `go_package` option must be specified in the `.proto` file. And in the command line, the `protoc-gen-go` plugin is deprecated, and the `--go-grpc_out` flag must be declared.
- The `UnimplementedProductInfoServer` must be embedded in the server struct definition.
- Add notes to explain what the `go_package` option is, why we need it and that we need to first push to the repository first in order to import the pb defintions
