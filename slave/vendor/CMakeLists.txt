include(ExternalProject)

if (APPLE)
  set(SO "dylib")
else ()
  set(SO "so")
endif()

#gtest
if(TESTS)
  set(GTEST_URL "https://github.com/google/googletest/archive/release-1.7.0.zip")
  if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/gtest-1.7.0.zip")
    set(GTEST_URL "${CMAKE_CURRENT_SOURCE_DIR}/gtest-1.7.0.zip")
  endif()

  ExternalProject_Add(gtest
    URL ${GTEST_URL}
    URL_MD5 "ef5e700c8a0f3ee123e2e0209b8b4961"
    DOWNLOAD_DIR ${CMAKE_CURRENT_BINARY_DIR}
    SOURCE_DIR ${CMAKE_CURRENT_BINARY_DIR}/gtest-1.7.0
    BUILD_IN_SOURCE 1
    BUILD_COMMAND "make"
    INSTALL_COMMAND ""
  )

  set(GTEST_INCLUDE_DIR "${CMAKE_CURRENT_BINARY_DIR}/gtest-1.7.0/include" PARENT_SCOPE)
  set(GTEST_BOTH_LIBRARIES "${CMAKE_CURRENT_BINARY_DIR}/gtest-1.7.0/libgtest.a" "${CMAKE_CURRENT_BINARY_DIR}/gtest-1.7.0/libgtest.a" PARENT_SCOPE)
endif(TESTS)

if(BENCH)
  set(CELERO_URL "https://github.com/DigitalInBlue/Celero/archive/v1.0.5.tar.gz")
  if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/v1.0.5.tar.gz")
    set(CELERO_URL "${CMAKE_CURRENT_SOURCE_DIR}/v1.0.5.tar.gz")
  endif()

  ExternalProject_Add(celero
    URL ${CELERO_URL}
    URL_MD5 "e2e19b8ebe89591e3e75efa1c16ed7bd"
    DOWNLOAD_DIR ${CMAKE_CURRENT_BINARY_DIR}
    SOURCE_DIR ${CMAKE_CURRENT_BINARY_DIR}/Celero
    CMAKE_ARGS "-DCELERO_RUN_EXAMPLE_ON_BUILD=NO"
    BUILD_IN_SOURCE 1
    INSTALL_COMMAND ""
  )

  set(CELERO_INCLUDE_DIR "${CMAKE_CURRENT_BINARY_DIR}/Celero/include" PARENT_SCOPE)
  set(CELERO_LIBRARY "${CMAKE_CURRENT_BINARY_DIR}/Celero/libcelero.${SO}" PARENT_SCOPE)
endif(BENCH)
