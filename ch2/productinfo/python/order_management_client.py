from __future__ import print_function
from endpoint import endpoint
from faker import Faker

import logging

import grpc
import copy
import json
import random

import order_management_pb2
import order_management_pb2_grpc

fake = Faker()


def make_pb_order(order):
    """
    Converts the JSON order into the protobuf format

    :param order:
    :return: the order in the protbuf format
    """
    return order_management_pb2.Order(id=order["id"], items=order["items"], description=order["description"], price=order["price"], destination=order["destination"])


def generate_updated_orders(orders: dict):
    """
    Select n numbers of orders where 1 <= n <= len(orders)
    and update some fields of the orders

    :param orders:
    :return: an iterator of orders
    """
    order_keys = list(orders.keys())
    num_orders_to_update = random.randint(1, len(order_keys))

    for _ in range(num_orders_to_update):
        rand_key_idx = random.randint(0, len(order_keys) - 1)
        key = order_keys[rand_key_idx]
        updated_order = copy.deepcopy(orders[key])
        updated_order["destination"] = fake.address()
        yield make_pb_order(updated_order)


def run():
    with grpc.insecure_channel(endpoint) as channel:
        stub = order_management_pb2_grpc.OrderManagementStub(channel)

        serialized_req = order_management_pb2.google_dot_protobuf_dot_wrappers__pb2.StringValue(
            value="1")
        get_order_res = stub.getOrder(serialized_req)
        print("Get order response {0}".format(get_order_res))

        # search the orders
        search_order_results = stub.searchOrders(order_management_pb2.SearchOrderQuery(query="Macbook"))
        for idx, search_order in enumerate(search_order_results):
            print("Search result {0}:\n".format(idx + 1))
            print(search_order)

        # update the orders - Client Streaming
        orders_json_data_path = "../data/example_orders.json"
        with open(orders_json_data_path) as orders_json:
            example_orders = json.load(orders_json)

        update_order_res = stub.updateOrders(generate_updated_orders(example_orders))
        print(update_order_res.value)

        proc_order_iterator = generate_updated_orders(example_orders)
        for shipment in stub.processOrders(proc_order_iterator):
            print(shipment)


if __name__ == "__main__":
    logging.basicConfig()
    run()
