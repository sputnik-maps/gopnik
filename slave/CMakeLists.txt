cmake_minimum_required(VERSION 2.8)

project(gopnikslave C CXX)

set(CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")

OPTION(TESTS "enable_testing" ON)
OPTION(BENCH "enable_benchmarks" ON)

include(CheckCXXSourceCompiles)

# Initialize FLAGS.
set(CMAKE_CXX_FLAGS "-Wall")
set(CMAKE_CXX_FLAGS_DEBUG "-O0 -g")
set(CMAKE_C_FLAGS_DEBUG "-O0 -g")
set(CMAKE_CXX_FLAGS_MINSIZEREL "-Os -DNDEBUG")
set(CMAKE_C_FLAGS_MINSIZEREL "-Os -DNDEBUG")
set(CMAKE_CXX_FLAGS_RELEASE "-O4 -DNDEBUG")
set(CMAKE_C_FLAGS_RELEASE "-O4 -DNDEBUG")
set(CMAKE_CXX_FLAGS_RELWITHDEBINFO "-O2 -g")
set(CMAKE_C_FLAGS_RELWITHDEBINFO "-O2 -g")

# Compiler-specific C++11 activation.
if ("${CMAKE_CXX_COMPILER_ID}" MATCHES "GNU")
	execute_process(
		COMMAND ${CMAKE_CXX_COMPILER} -dumpversion OUTPUT_VARIABLE GCC_VERSION)
	if (NOT (GCC_VERSION VERSION_GREATER 4.7 OR GCC_VERSION VERSION_EQUAL 4.7))
		message(FATAL_ERROR "${PROJECT_NAME} requires g++ 4.7 or greater.")
	endif ()
	set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=gnu++11")
elseif ("${CMAKE_CXX_COMPILER_ID}" MATCHES "Clang")
	set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11")
else ()
	message(FATAL_ERROR "Your C++ compiler does not support C++11.")
endif ()
message(STATUS "CMAKE_CXX_FLAGS: ${CMAKE_CXX_FLAGS}")

# ICU
find_package(ICU REQUIRED)
include_directories(${ICU_INCLUDE_DIRS})

# Setup boost
find_package( Boost 1.48.0 COMPONENTS system filesystem program_options REQUIRED )
include_directories(${Boost_INCLUDE_DIRS})

# Threads
find_package(Threads REQUIRED)

# Protobuf
find_package(Protobuf REQUIRED)
include_directories(${PROTOBUF_INCLUDE_DIR})

#Freetype
find_package(Freetype REQUIRED)
include_directories(${FREETYPE_INCLUDE_DIRS})

# Cairo
find_package(Cairo REQUIRED)
include_directories(${CAIRO_INCLUDE_DIRS})

# Proj
find_package(Proj REQUIRED)
include_directories(${PROJ_INCLUDE_DIR})

# Setup mapnik
find_package( Mapnik REQUIRED)
include_directories(${MAPNIK_INCLUDE_DIRS})
set(CMAKE_MODULE_LINKER_FLAGS ${MAPNIK_LDFLAGS})
set(CMAKE_SHARED_LINKER_FLAGS ${MAPNIK_LDFLAGS})

# "Vendor" staff
add_subdirectory(vendor)

# Protobuf code
add_subdirectory(proto)

# Code
add_subdirectory(src)

# Set up sampledata

if (TESTS OR BENCH)
	add_subdirectory(sampledata)
endif (TESTS OR BENCH)

# GTest magic
if(TESTS)
	message(STATUS "GTest include dir: ${GTEST_INCLUDE_DIR}")
	include_directories(${GTEST_INCLUDE_DIR})
	add_custom_target(test)
	macro(run_test test_target)
		add_custom_target(${test_target}_runtest
			COMMAND ${test_target}
			DEPENDS ${test_target}
			WORKING_DIRECTORY "${CMAKE_BINARY_DIR}")
		add_dependencies(test ${test_target}_runtest)
	endmacro()
	add_subdirectory(tests)
endif(TESTS)

# Celero magic
if(BENCH)
	message(STATUS "Celero include dir: ${CELERO_INCLUDE_DIR}")
	include_directories(${CELERO_INCLUDE_DIR})
	add_custom_target(bench)
	macro(run_bench bench_target)
		add_custom_target(${bench_target}_runbench
			COMMAND ${bench_target}
			DEPENDS ${bench_target}
			WORKING_DIRECTORY "${CMAKE_BINARY_DIR}")
		add_dependencies(bench ${bench_target}_runbench)
	endmacro()
	add_subdirectory(bench)
endif(BENCH)
