python -m grpc_tools.protoc -I../../../../protos/productinfo --python_out=. --grpc_python_out=. ../../../../protos/productinfo/product_info.proto
python -m grpc_tools.protoc -I../../../../protos/order_management --python_out=. --grpc_python_out=. ../../../../protos/order_management/order_management.proto
