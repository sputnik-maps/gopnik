#pragma once

#include "render.hpp"

#include <string>
#include <vector>
#include <memory>

namespace gopnik {

class TestRenderImpl : public Render {
	public:
		TestRenderImpl(int sleepTime);
		virtual ~TestRenderImpl();

		virtual std::shared_ptr<Result> Do(Task const& task);
	private:
		int sleepTime_;
};

}
