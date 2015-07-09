#include "loop.hpp"
#include "test_render_impl.hpp"

#include <iostream>
#include <stdexcept>
#include <cstdint>

#include <boost/program_options.hpp>

namespace options = boost::program_options;
using namespace gopnik;

int
main(int argc, char *argv[]) {
	try {
		// Parse args
		int sleepTime = -1;
		options::options_description desc (std::string (argv[0]).append(" options"));
		desc.add_options()
			("help", "Display this message")
			("sleep", options::value<int>(), "sleep in seconds")
		;
		options::variables_map args;
		options::store (options::command_line_parser (argc, argv).options (desc)
			.style (options::command_line_style::default_style |
				options::command_line_style::allow_long_disguise)
			.run (), args);
		options::notify (args);

		if (args.count("h")) {
			std::cerr << desc << std::endl;
			return 1;
		}

		if (args.count("sleep")) {
			sleepTime = args["sleep"].as<int>();
		}

		// Set up render
		TestRenderImpl render{sleepTime};

		Loop loop{0 /* stdin */, 1 /* stdout */, render};

		// Start event loop
		loop.Run();
	}
	catch(std::exception const& e) {
		std::cerr << "Exception: " << e.what() << std::endl;
		return 1;
	}
	catch(...) {
		std::cerr << "Unknown exception" << std::endl;
		return 2;
	}
	return 0;
}
