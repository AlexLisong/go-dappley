#include <v8.h>

using namespace v8;

void NewBlacklistInstance(Isolate *isolate, Local<Context> context, void *address);
void blacklistCallback(const FunctionCallbackInfo<Value> &info);