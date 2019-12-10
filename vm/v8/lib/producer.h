#include <v8.h>

using namespace v8;

void NewProducerInstance(Isolate *isolate, Local<Context> context, void *address);
void producerCallback(const FunctionCallbackInfo<Value> &info);
