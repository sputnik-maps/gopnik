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

service Render extends base_service.BaseService {
	list<types.Tile> render(1: types.Coord coord, 2: Priority prio, 3: bool wait_storage) throws (1: RenderError renderErr, 2: QueueLimitExceeded queueErr),
}
