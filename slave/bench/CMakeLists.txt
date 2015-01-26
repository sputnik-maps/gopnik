include_directories("${PROJECT_SOURCE_DIR}/src")
include_directories(${GOPNIK_SAMPLEDATA_INCLUDE_DIR})
include_directories(${PROTO_DIR})

file(GLOB GOPNIKSLAVE_BENCH_SOURCES "*.cpp")
add_executable(gopnikslave_bench ${GOPNIKSLAVE_BENCH_SOURCES})

add_dependencies(gopnikslave_bench
  gopnikslave_lib
  celero
)

target_link_libraries(gopnikslave_bench
  gopnikslave_lib
  ${CELERO_LIBRARY}
)

run_bench(gopnikslave_bench)

