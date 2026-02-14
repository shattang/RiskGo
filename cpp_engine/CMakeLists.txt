cmake_minimum_required(VERSION 3.10)
project(RiskEngine)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 1. Threads is standard
find_package(Threads REQUIRED)

# 2. Setup PkgConfig as a robust fallback
find_package(PkgConfig REQUIRED)

# --- Robust Dependency Discovery ---

pkg_check_modules(QuantLib REQUIRED quantlib)
pkg_check_modules(Protobuf REQUIRED protobuf)
pkg_check_modules(gRPC REQUIRED grpc++)

# Find Protobuf compiler and gRPC plugin
find_program(PROTOC_EXECUTABLE NAMES protoc)
find_program(GRPC_CPP_PLUGIN_EXECUTABLE NAMES grpc_cpp_plugin)

if(NOT PROTOC_EXECUTABLE OR NOT GRPC_CPP_PLUGIN_EXECUTABLE)
    message(FATAL_ERROR "Could not find protoc or grpc_cpp_plugin")
endif()

include_directories(
    SYSTEM
    ${QuantLib_INCLUDE_DIRS}
    ${Protobuf_INCLUDE_DIRS}
    ${gRPC_INCLUDE_DIRS}
)

# Protobuf and gRPC generation
set(PROTO_SRC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../protos)
set(PROTO_OUT_DIR ${CMAKE_CURRENT_BINARY_DIR}/generated)
file(MAKE_DIRECTORY ${PROTO_OUT_DIR})

set(PROTO_FILE ${PROTO_SRC_DIR}/risk_engine.proto)

add_custom_command(
    OUTPUT "${PROTO_OUT_DIR}/risk_engine.pb.cc"
           "${PROTO_OUT_DIR}/risk_engine.pb.h"
           "${PROTO_OUT_DIR}/risk_engine.grpc.pb.cc"
           "${PROTO_OUT_DIR}/risk_engine.grpc.pb.h"
    COMMAND ${PROTOC_EXECUTABLE}
    ARGS --cpp_out=${PROTO_OUT_DIR}
         --grpc_out=${PROTO_OUT_DIR}
         --plugin=protoc-gen-grpc=${GRPC_CPP_PLUGIN_EXECUTABLE}
         -I${PROTO_SRC_DIR}
         ${PROTO_FILE}
    DEPENDS ${PROTO_FILE}
)

include_directories(${PROTO_OUT_DIR})

add_executable(risk_engine_server
    src/server.cpp
    "${PROTO_OUT_DIR}/risk_engine.pb.cc"
    "${PROTO_OUT_DIR}/risk_engine.grpc.pb.cc"
)

target_link_libraries(risk_engine_server
    ${gRPC_LIBRARIES}
    ${Protobuf_LIBRARIES}
    ${QuantLib_LIBRARIES}
    Threads::Threads
)

message(STATUS "QuantLib_LIBRARIES: ${QuantLib_LIBRARIES}")
message(STATUS "gRPC_LIBRARIES: ${gRPC_LIBRARIES}")
message(STATUS "Protobuf_LIBRARIES: ${Protobuf_LIBRARIES}")
