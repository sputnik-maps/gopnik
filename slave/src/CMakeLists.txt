include_directories(${PROTO_DIR})
add_library(gopnikslave_lib STATIC
	loop.cpp
	render.cpp
	render_impl.cpp
)
target_link_libraries(gopnikslave_lib
	${MAPNIK_LIBRARIES}
	${PROTO_LIB}
	${FREETYPE_LIBRARIES}
	${Boost_LIBRARIES}
	${PROJ_LIBRARY}
	${ICU_LIBRARIES}
)

add_executable(gopnikslave main.cpp)
add_dependencies(gopnikslave gopnikslave_lib)
target_link_libraries(gopnikslave
	gopnikslave_lib
	${GFLAGS_LIBRARIES}
)

install(TARGETS gopnikslave DESTINATION ${PROJECT_SOURCE_DIR}/../bin)
