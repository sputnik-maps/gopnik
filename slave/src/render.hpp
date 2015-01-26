#pragma once

#include "msg.pb.h"

#include <memory>

namespace gopnik {

class Render {
	public:
		virtual ~Render();

		virtual std::shared_ptr<Result> Do(Task const& task) = 0;
};

}
