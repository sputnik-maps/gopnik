# - Try to find Mapnik
# Once done this will define
#
#  MAPNIK_FOUND - system has Mapnik
#  MAPNIK_INCLUDE_DIRS - the Mapnik include directory
#  MAPNIK_LIBRARIES - Link these to use Mapnik
#  MAPNIK_CONFIG - mapnik-config binary
#  MAPNIK_CXXFLAGS - mapnik-config --cflags)
#  MAPNIK_LDFLAGS - mapnik-config --libs)
#  MAPNIK_PLUGINDIR - mapnik-config --input-plugins)
#
#  Copyright (c) 2007 Andreas Schneider <mail@cynapses.org>
#  Copyright (c) 2014 Maxim Dementyev <max@orofarne.ru>
#
#  Redistribution and use is allowed according to the terms of the New
#  BSD license.
#  For details see the accompanying COPYING-CMAKE-SCRIPTS file.
#


if (MAPNIK_LIBRARIES AND MAPNIK_INCLUDE_DIRS)
  # in cache already
  set(MAPNIK_FOUND TRUE)
else (MAPNIK_LIBRARIES AND MAPNIK_INCLUDE_DIRS)
  find_path(MAPNIK_INCLUDE_DIR
    NAMES
      mapnik/config.hpp
    PATHS
      /usr/include
      /usr/local/include
      /opt/local/include
      /sw/include
  )

  find_library(MAPNIK_LIBRARY
    NAMES
      mapnik
      mapnik2
    PATHS
      /usr/lib
      /usr/local/lib
      /opt/local/lib
      /sw/lib
  )

  if (MAPNIK_LIBRARY)
    set(MAPNIK_FOUND TRUE)
  endif (MAPNIK_LIBRARY)

  set(MAPNIK_INCLUDE_DIRS
    ${MAPNIK_INCLUDE_DIR}
  )

  if (MAPNIK_FOUND)
    set(MAPNIK_LIBRARIES
      ${MAPNIK_LIBRARIES}
      ${MAPNIK_LIBRARY}
    )
  endif (MAPNIK_FOUND)

  if (MAPNIK_INCLUDE_DIRS AND MAPNIK_LIBRARIES)
     set(MAPNIK_FOUND TRUE)
  endif (MAPNIK_INCLUDE_DIRS AND MAPNIK_LIBRARIES)

  if (MAPNIK_FOUND)
    if (NOT Mapnik_FIND_QUIETLY)
      message(STATUS "Found Mapnik: ${MAPNIK_LIBRARIES}")
    endif (NOT Mapnik_FIND_QUIETLY)
  else (MAPNIK_FOUND)
    if (Mapnik_FIND_REQUIRED)
      message(FATAL_ERROR "Could not find Mapnik")
    endif (Mapnik_FIND_REQUIRED)
  endif (MAPNIK_FOUND)

  find_program(MAPNIK_CONFIG
    NAMES
      mapnik-config
    PATHS
      /usr/bin
      /usr/local/bin
      /opt/local/bin
  )

  if (MAPNIK_CONFIG)
    message(STATUS "Found mapnik-config: ${MAPNIK_CONFIG}")
    execute_process(COMMAND mapnik-config --cflags OUTPUT_VARIABLE MAPNIK_CXXFLAGS OUTPUT_STRIP_TRAILING_WHITESPACE)
    execute_process(COMMAND mapnik-config --libs OUTPUT_VARIABLE MAPNIK_LDFLAGS OUTPUT_STRIP_TRAILING_WHITESPACE)
    execute_process(COMMAND mapnik-config --input-plugins OUTPUT_VARIABLE MAPNIK_PLUGINDIR OUTPUT_STRIP_TRAILING_WHITESPACE)
    message(STATUS "Mapnik default plugin path: ${MAPNIK_PLUGINDIR}")
    execute_process(COMMAND mapnik-config --dep-includes OUTPUT_VARIABLE MAPNIK_INCLUDE_FLAGS OUTPUT_STRIP_TRAILING_WHITESPACE)
    string(REPLACE " " ";" MAPNIK_INCLUDE_FLAG_LIST ${MAPNIK_INCLUDE_FLAGS})
    foreach(INCLUDEFLAG ${MAPNIK_INCLUDE_FLAG_LIST})
        string(REPLACE "-I" "" INCLUDEPATH ${INCLUDEFLAG})
        set(MAPNIK_INCLUDE_DIRS ${MAPNIK_INCLUDE_DIRS} ${INCLUDEPATH})
    endforeach(INCLUDEFLAG)
  endif (MAPNIK_CONFIG)

  # show the variables only in the advanced view
  mark_as_advanced(
      MAPNIK_INCLUDE_DIRS
      MAPNIK_LIBRARIES
      MAPNIK_CXXFLAGS
      MAPNIK_LDFLAGS
      MAPNIK_PLUGINDIR
    )

endif (MAPNIK_LIBRARIES AND MAPNIK_INCLUDE_DIRS)

