#pragma once

#include "render.hpp"

#include <memory>

namespace gopnik {

class Loop {
	public:
		Loop(int in_fd, int out_fd, Render &render);
		virtual ~Loop();

		void Run();
	protected:
		class impl;
		std::unique_ptr<impl> pimpl_;

		void processOne();
};

}
