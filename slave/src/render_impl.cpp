#include "render_impl.hpp"

#include <algorithm>
#include <cstdint>
#include <stdexcept>
#include <iostream>

#include <boost/filesystem.hpp>
#include <boost/static_assert.hpp>
#include <boost/format.hpp>

#include <mapnik/version.hpp>
#include <mapnik/map.hpp>
#include <mapnik/datasource_cache.hpp>
#include <mapnik/agg_renderer.hpp>
#include <mapnik/load_map.hpp>
#include <mapnik/graphics.hpp>
#include <mapnik/image_util.hpp>
#include <mapnik/box2d.hpp>

#include <proj_api.h>

BOOST_STATIC_ASSERT(MAPNIK_VERSION >= 200200);

namespace gopnik {

class x_pj_free {
	public:
		void operator()(projPJ pj) const {
			if (pj) {
				pj_free(pj);
			}
		}
};

class RenderImpl::impl {
	public:
		std::unique_ptr<mapnik::Map> map_;
		unsigned tile_size_;
		int buffer_size_;
		double scale_factor_;
		std::unique_ptr<void, x_pj_free> pj_source_;
		std::unique_ptr<void, x_pj_free> pj_target_;
	public:
		impl(std::string stylesheet, std::vector<std::string> fonts_path, std::string plugins_path, unsigned tile_size, int buffer_size, double scale_factor);
		void appendTile(std::shared_ptr<Result> res, mapnik::image_view<mapnik::image_data_32> vw);
		mapnik::box2d<double> getBbox(Task const &task);
	protected:
		void loadFonts(std::string path);
		void analyzeTile(Tile *tile, mapnik::image_view<mapnik::image_data_32> vw);
		void convertPoint(double &x, double &y, int zoom);
};

RenderImpl::RenderImpl(std::string stylesheet, std::vector<std::string> fonts_path, std::string plugins_path, unsigned tile_size, int buffer_size, double scale_factor)
	: pimpl_(new impl{stylesheet, fonts_path, plugins_path, tile_size, buffer_size, scale_factor})
{
}

RenderImpl::impl::impl(std::string stylesheet, std::vector<std::string> fonts_path, std::string plugins_path, unsigned tile_size, int buffer_size, double scale_factor)
	: tile_size_(tile_size)
	, buffer_size_(buffer_size)
	, scale_factor_(scale_factor)
{
	mapnik::datasource_cache::instance().register_datasources(plugins_path);
	for (auto it = fonts_path.begin(); it != fonts_path.end(); ++it) {
		loadFonts(*it);
	}
	map_.reset(new mapnik::Map{});
	mapnik::load_map(*map_, stylesheet);
	pj_source_.reset(pj_init_plus(mapnik::MAPNIK_GMERC_PROJ.c_str()));
	if (!pj_source_) {
		throw std::runtime_error("Invalid source projection");
	}
	pj_target_.reset(pj_init_plus(map_->srs().c_str()));
	if (!pj_target_) {
		throw std::runtime_error("Invalid destination projection");
	}
}

void
RenderImpl::impl::loadFonts(std::string fonts_path) {
	using namespace boost::filesystem;

	path p(fonts_path);
	if (exists(p)) {
		if (is_regular_file(p) && p.extension() == ".ttf") {
			mapnik::freetype_engine::register_font(p.string());
		} else if (is_directory(p)) {
			std::for_each(directory_iterator(p), directory_iterator(), [this](directory_entry &ent) {
				loadFonts(ent.path().string());
			});
		}
	}
}


RenderImpl::~RenderImpl() {

}

void
RenderImpl::impl::convertPoint(double &x, double &y, int zoom) {
	// Web mercator
	static const double bound_x0 = -20037508.3428;
	static const double bound_x1 =  20037508.3428;
	static const double bound_y0 = -20037508.3428;
	static const double bound_y1 =  20037508.3428;

	x = bound_x0 + (bound_x1 - bound_x0)* ((double)x / (double)(1<<zoom));
	y = bound_y1 - (bound_y1 - bound_y0)* ((double)y / (double)(1<<zoom));

	auto err = pj_transform(pj_source_.get(), pj_target_.get(), 1, 1, &x, &y, NULL);
	if (err) {
		throw std::runtime_error(
				(boost::format("pj_transform error %1%: %2%") % err % pj_strerrno(err)).str()
			);
	}
}

mapnik::box2d<double>
RenderImpl::impl::getBbox(Task const &task) {
	double x1, y1, x2, y2;

	x1 = task.x();
	y1 = task.y();
	x2 = task.x() + task.size();
	y2 = task.y() + task.size();

	convertPoint(x1, y1, task.zoom());
	convertPoint(x2, y2, task.zoom());

	return mapnik::box2d<double>(x1, y1, x2, y2);
}

std::shared_ptr<Result>
RenderImpl::Do(Task const &task) {
	std::shared_ptr<Result> res(new Result{});

	// Configure map
	pimpl_->map_->resize(pimpl_->tile_size_ * task.size(), pimpl_->tile_size_ * task.size());
	auto bbox = pimpl_->getBbox(task);
	pimpl_->map_->zoom_to_box(bbox);
	if (pimpl_->buffer_size_ >= 0) {
		pimpl_->map_->set_buffer_size(pimpl_->buffer_size_);
	}

	// Render map
	mapnik::image_32 buf(pimpl_->map_->width(), pimpl_->map_->height());
	mapnik::agg_renderer<mapnik::image_32> ren(*pimpl_->map_, buf, pimpl_->scale_factor_);
	ren.apply();

	// Split the meta tile into an NxN grid of tiles
	unsigned x, y;
	for (y = 0; y < task.size(); ++y) {
		for (x = 0; x < task.size(); ++x) {
			mapnik::image_view<mapnik::image_data_32> vw{
					x * pimpl_->tile_size_, y * pimpl_->tile_size_,
					pimpl_->tile_size_, pimpl_->tile_size_,
					buf.data()};
			pimpl_->appendTile(res, vw);
		}
	}

	return res;
}

void
RenderImpl::impl::appendTile(std::shared_ptr<Result> res, mapnik::image_view<mapnik::image_data_32> vw) {
	Tile *tile = res->add_tiles();

	// Encode tile
	std::string tile_png = mapnik::save_to_string(vw, "png256");
	tile->set_png(tile_png);

	// Analyze tile
	analyzeTile(tile, vw);
}

void
RenderImpl::impl::analyzeTile(Tile *tile, mapnik::image_view<mapnik::image_data_32> vw) {
	// Analyze tile
	unsigned x, y;
	mapnik::image_data_32::pixel_type prev_pixel;
	for (y = 0; y < vw.height(); ++y) {
		mapnik::image_data_32::pixel_type const *row = vw.getRow(y);
		for (x = 0; x < vw.width(); ++x) {
			mapnik::image_data_32::pixel_type pixel = row[x];
			if(x > 0 && y > 0 && pixel != prev_pixel) {
				return;
			}
			prev_pixel = pixel;
		}
	}

	// Ok, it's single-color tile
	Color *col = tile->mutable_single_color();
	col->set_r(prev_pixel & 0xff);
	col->set_g((prev_pixel >> 8 ) & 0xff);
	col->set_b((prev_pixel >> 16) & 0xff);
	col->set_a((prev_pixel >> 24) & 0xff);
}

}
