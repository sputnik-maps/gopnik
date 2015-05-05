#include "loop.hpp"

#include <unistd.h>
#include "portable_endian.h"

#include <cstdint>
#include <vector>

#include "msg.pb.h"

namespace gopnik {

// Messege struct:
//
//   +------------------+
//   |  uint64_t size   |
//   |  (LittleEndian)  |
//   +------------------+
//   |                  |
//   |    Protobuf      |
//   | encoded message  |
//   |                  |
//   +------------------+
//

class Loop::impl {
	public:
		int in_fd;
		int out_fd;
		Render &render;

		impl(int in_fd, int out_fd, Render &rend)
			: in_fd(in_fd)
			, out_fd(out_fd)
			, render(rend)
		{
		}

		void readMessage(std::string &data);
		void writeMessage(std::string &data);
		void readUint(uint64_t &val);
		void writeUint(uint64_t val);
};

Loop::Loop(int in_fd, int out_fd, Render &render)
	: pimpl_(new impl(in_fd, out_fd, render))
{
}

Loop::~Loop() {
}

void
Loop::Run() {
	// Hello message
	pimpl_->writeUint(0);

	// Event loop
	for(;;) {
		processOne();
	}
}

void
Loop::processOne() {
	// Read task
	Task task;
	std::string in_buf;
	pimpl_->readMessage(in_buf);
	task.ParseFromString(in_buf);

	// Do work
	std::shared_ptr<Result> res = pimpl_->render.Do(task);

	// Write answer
	std::string out_buf;
	res->SerializeToString(&out_buf);
	pimpl_->writeMessage(out_buf);
}

void
Loop::impl::readMessage(std::string &data) {
	uint64_t message_size;
	readUint(message_size);
	data.resize(message_size);
	int rc = ::read(in_fd, const_cast<char *>(data.data()), message_size);
	if (rc < 0 || uint64_t(rc) != message_size) {
		throw std::runtime_error("Invalid read");
	}
}

void
Loop::impl::writeMessage(std::string &data) {
	writeUint(data.size());
	int rc = ::write(out_fd, data.data(), data.size());
	if (rc < 0 || uint64_t(rc) != data.size()) {
		throw std::runtime_error("Invalid write");
	}
}

void
Loop::impl::readUint(uint64_t &val) {
	int rc = ::read(in_fd, reinterpret_cast<void *>(&val), sizeof(uint64_t));
	if (rc != sizeof(uint64_t)) {
		throw std::runtime_error("invalid read");
	}
	val = le64toh(val);
}

void
Loop::impl::writeUint(uint64_t val) {
	val = htole64(val);
	int rc = ::write(out_fd, reinterpret_cast<void *>(&val), sizeof(uint64_t));
	if (rc != sizeof(uint64_t)) {
		throw std::runtime_error("invalid write");
	}
}


}
