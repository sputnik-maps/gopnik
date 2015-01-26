set(proto_packages
	${CMAKE_CURRENT_SOURCE_DIR}/msg.proto
)
set(proto_packages_cpp
	${CMAKE_CURRENT_BINARY_DIR}/msg.pb.h
	${CMAKE_CURRENT_BINARY_DIR}/msg.pb.cc
)

add_custom_command(
	OUTPUT ${proto_packages_cpp}
	DEPENDS ${proto_packages}
	COMMAND ${PROTOBUF_PROTOC_EXECUTABLE} --proto_path=${CMAKE_CURRENT_SOURCE_DIR} --cpp_out=${CMAKE_CURRENT_BINARY_DIR} ${proto_packages}
)

add_custom_target(protobuf_autogeneration_target ALL
	DEPENDS ${proto_packages}
)

include_directories(${CMAKE_CURRENT_BINARY_DIR})
add_library(gopnikslave_proto_lib STATIC
	${proto_packages_cpp}
)
add_dependencies(gopnikslave_proto_lib protobuf_autogeneration_target)
target_link_libraries(gopnikslave_proto_lib
	${PROTOBUF_LIBRARIES}
)

set(PROTO_LIB gopnikslave_proto_lib PARENT_SCOPE)
set(PROTO_DIR ${CMAKE_CURRENT_BINARY_DIR} PARENT_SCOPE)
