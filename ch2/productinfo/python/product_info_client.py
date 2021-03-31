from __future__ import print_function
from endpoint import endpoint
import logging

import grpc

import product_info_pb2
import product_info_pb2_grpc


def run():
    with grpc.insecure_channel(endpoint) as channel:
        stub = product_info_pb2_grpc.ProductInfoStub(channel)
        # example product info
        name = "Apple iPhone 10 Max"
        desc = "iPhone 10 Max Triple Core Camera"
        price = 2400

        res_one = stub.addProduct(product_info_pb2.Product(name=name, description=desc, price=price))
        print(f"Product ID: {res_one.value} added successfully\n")

        res_two = stub.getProduct(product_info_pb2.ProductID(value=res_one.value))
        print(f"Product ID {res_one.value}:\n\n{res_two}")


if __name__ == "__main__":
    logging.basicConfig()
    run()
