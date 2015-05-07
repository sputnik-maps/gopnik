namespace go gopnikrpc
namespace cpp gopnikrpc

include "types.thrift"
include "base_service.thrift"

enum Priority {
	HIGH = 1,
	LOW = 2,
}

exception RenderError {
	1: string message
}

exception QueueLimitExceeded {
}

struct RenderResponse {
	1: list<types.Tile> tiles,
	2: i64 renderTime, // Nanoseconds
	3: i64 saveTime, // Nanoseconds
}

service Render extends base_service.BaseService {
	RenderResponse render(1: types.Coord coord, 2: Priority prio, 3: bool wait_storage) throws (1: RenderError renderErr, 2: QueueLimitExceeded queueErr),
}
