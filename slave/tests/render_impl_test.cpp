#include "render_impl.hpp"
#include "sampledata.h"

#include <gtest/gtest.h>

#include <fstream>

#include <boost/filesystem.hpp>

using namespace gopnik;

TEST(Render, GenTile) {
	boost::filesystem::current_path(GOPNIK_SAMPLE_DATA_PATH);

	RenderImpl render{"stylesheet.xml", std::vector<std::string>(), MAPNIK_PLUGINDIR, 256, 128, 1.0, "png256"};
	Task task;
	task.set_x(0);
	task.set_y(0);
	task.set_zoom(1);
	task.set_size(1);

	auto res = render.Do(task);
	ASSERT_EQ(1, res->tiles_size());
	Tile const &tile = res->tiles(0);
	EXPECT_FALSE(tile.has_single_color());
	auto data = tile.png();
	ASSERT_GT(data.size(), 0);

	// std::ofstream f( "/tmp/test.png" );
	// f.write(data.data(), data.size());
}

TEST(Render, Gen10Tiles) {
	boost::filesystem::current_path(GOPNIK_SAMPLE_DATA_PATH);

	RenderImpl render{"stylesheet.xml", std::vector<std::string>(), MAPNIK_PLUGINDIR, 256, 128, 1.0, "png256"};
	Task task;
	task.set_x(0);
	task.set_y(0);
	task.set_zoom(1);
	task.set_size(1);

	for (int i = 0; i < 10; ++i) {
		auto res = render.Do(task);
		ASSERT_EQ(1, res->tiles_size());
	}
}

TEST(Render, GenMetaTile1z) {
	boost::filesystem::current_path(GOPNIK_SAMPLE_DATA_PATH);

	RenderImpl render{"stylesheet.xml", std::vector<std::string>(), MAPNIK_PLUGINDIR, 256, 128, 1.0, "png256"};
	Task task;
	task.set_x(0);
	task.set_y(0);
	task.set_zoom(1);
	task.set_size(4);

	auto res = render.Do(task);
	ASSERT_EQ(task.size()*task.size(), res->tiles_size());
	Tile const &tile = res->tiles(0);
	EXPECT_FALSE(tile.has_single_color());
	auto data = tile.png();
	ASSERT_GT(data.size(), 0);

	// std::ofstream f( "/tmp/test.png" );
	// f.write(data.data(), data.size());
}

TEST(Render, GenMetaTile) {
	boost::filesystem::current_path(GOPNIK_SAMPLE_DATA_PATH);

	RenderImpl render{"stylesheet.xml", std::vector<std::string>(), MAPNIK_PLUGINDIR, 256, 128, 1.0, "png256"};
	Task task;
	task.set_x(33);
	task.set_y(77);
	task.set_zoom(14);
	task.set_size(8);

	auto res = render.Do(task);
	ASSERT_EQ(64, res->tiles_size());
}
