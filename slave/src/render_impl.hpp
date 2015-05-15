#pragma once

#include "render.hpp"

#include <string>
#include <vector>
#include <memory>

namespace gopnik {

class RenderImpl : public Render {
	public:
		RenderImpl(std::string stylesheet, std::vector<std::string> fonts_path, std::string plugins_path, unsigned tile_size, int buffer_size, double scale_factor, std::string const& image_format);
		virtual ~RenderImpl();

		virtual std::shared_ptr<Result> Do(Task const& task);

	private:
		class impl;
		std::unique_ptr<impl> pimpl_;

};

}
