from __future__ import print_function
from endpoint import endpoint
import logging

import grpc

import order_management_pb2
import order_management_pb2_grpc


def run():
    with grpc.insecure_channel(endpoint) as channel:
        stub = order_management_pb2_grpc.OrderManagementStub(channel)

        serialized_req = order_management_pb2.google_dot_protobuf_dot_wrappers__pb2.StringValue(
                value="iPod Air Order")
        get_order_res = stub.getOrder(serialized_req)
        print("Get order response " + get_order_res)

        search_order_results = stub.searchOrders(order_management_pb2.SearchOrderQuery(query="Ma"))
        for search_order in search_order_results:
            print(search_order)


if __name__ == "__main__":
    logging.basicConfig()
    run()
