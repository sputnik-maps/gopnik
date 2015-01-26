#include "loop.hpp"
#include "render_impl.hpp"

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
		std::string stylesheet;
		std::vector<std::string> fontsPath;
		std::string pluginsPath;
		unsigned tileSize = 256;
		int bufferSize = -1;
		double scaleFactor = 1.0;

		options::options_description desc (std::string (argv[0]).append(" options"));
		desc.add_options()
			("help", "Display this message")
			("stylesheet", options::value<std::string>(), "stylesheet")
			("tileSize", options::value<unsigned>(), "tileSize")
			("bufferSize", options::value<int>(), "tileSize")
			("fontsPath", options::value<std::vector<std::string>>(), "fontsPath")
			("pluginsPath", options::value<std::string>(), "pluginsPath")
			("scaleFactor", options::value<double>(), "scaleFactor")
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
		if (args.count("stylesheet")) {
			stylesheet = args["stylesheet"].as<std::string>();
		} else {
			throw std::runtime_error("stylesheet required");
		}
		if (args.count("tileSize")) {
			tileSize = args["tileSize"].as<unsigned>();
		}
		if (args.count("bufferSize")) {
			bufferSize = args["bufferSize"].as<int>();
		}
		if (args.count("fontsPath")) {
			fontsPath = args["fontsPath"].as<std::vector<std::string>>();
		}
		if (args.count("pluginsPath")) {
			pluginsPath = args["pluginsPath"].as<std::string>();
		}
		if (args.count("scaleFactor")) {
			scaleFactor = args["scaleFactor"].as<double>();
		}

		// Set up render
		RenderImpl render{stylesheet, fontsPath, pluginsPath,
			tileSize, bufferSize, scaleFactor};

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
