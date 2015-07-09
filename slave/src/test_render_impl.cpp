#include "test_render_impl.hpp"

#include <unistd.h>

namespace gopnik {

TestRenderImpl::TestRenderImpl(int sleepTime)
    :sleepTime_(sleepTime)
{
}

TestRenderImpl::~TestRenderImpl(){
}

std::shared_ptr<Result> TestRenderImpl::Do(Task const& task){
    sleep(sleepTime_);
    return std::shared_ptr<Result>();
}

} //namespace