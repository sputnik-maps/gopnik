include_directories("${PROJECT_SOURCE_DIR}/src")
include_directories(${GOPNIK_SAMPLEDATA_INCLUDE_DIR})
include_directories(${PROTO_DIR})

file(GLOB GOPNIKSLAVE_TEST_SOURCES "*.cpp")
add_executable(gopnikslave_test ${GOPNIKSLAVE_TEST_SOURCES})

add_dependencies(gopnikslave_test
	gtest
)

target_link_libraries(gopnikslave_test
	gopnikslave_lib
	${GTEST_BOTH_LIBRARIES}
	${CMAKE_THREAD_LIBS_INIT}
	${Boost_LIBRARIES}
	"dl"
)

run_test(gopnikslave_test)
