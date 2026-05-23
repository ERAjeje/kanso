#!/bin/bash
python -m grpc_tools.protoc \
  -Iproto \
  --python_out=src \
  --grpc_python_out=src \
  proto/analysis.proto
sed -i 's/import analysis_pb2/from . import analysis_pb2/' src/analysis_pb2_grpc.py
