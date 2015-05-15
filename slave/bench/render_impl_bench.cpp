#include <celero/Celero.h>

#include "render_impl.hpp"
#include "sampledata.h"

#include <memory>
#include <iostream>

#include <boost/filesystem.hpp>

using namespace gopnik;

class RenderImplBenchFixture : public celero::TestFixture {
	public:
		virtual void setUp(int64_t experimentValue) {
			boost::filesystem::current_path(GOPNIK_SAMPLE_DATA_PATH);
			render_.reset(new RenderImpl{"stylesheet.xml", std::vector<std::string>(), MAPNIK_PLUGINDIR, 256, 128, 1.0, "png256"});
		}

		virtual void tearDown() {
			render_.reset();
		}

		std::unique_ptr<RenderImpl> render_;
};

BASELINE_F(RenderImplBenchTest, Baseline, RenderImplBenchFixture, 10, 100) {
	Task task;
	task.set_x(0);
	task.set_y(0);
	task.set_zoom(1);
	task.set_size(1);
	celero::DoNotOptimizeAway(task);
}

BENCHMARK_F(RenderImplBenchTest, Hello, RenderImplBenchFixture, 10, 100) {
	Task task;
	task.set_x(0);
	task.set_y(0);
	task.set_zoom(1);
	task.set_size(1);

	auto res = render_->Do(task);
	celero::DoNotOptimizeAway(res);
}
