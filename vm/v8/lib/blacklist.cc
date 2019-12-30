#include "blacklist.h"
#include "../engine.h"
#include "memory.h"

static FuncBlacklist fBlacklist = NULL;

void InitializeBlacklist(FuncBlacklist blacklist) {
    fBlacklist = blacklist;
}

void NewBlacklistInstance(Isolate *isolate, Local<Context> context, void *address) {
    Local<ObjectTemplate> blacklistTpl = ObjectTemplate::New(isolate);
   blacklistTpl->SetInternalFieldCount(1);

    blacklistTpl->Set(String::NewFromUtf8(isolate, "blacklist"), FunctionTemplate::New(isolate, blacklistCallback),
                    static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));

    Local<Object> instance = blacklistTpl->NewInstance(context).ToLocalChecked();
    instance->SetInternalField(0, External::New(isolate, address));
    context->Global()->DefineOwnProperty(context, String::NewFromUtf8(isolate, "_native_blacklist"), instance,
                                         static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));
}

// blacklistCallback
void blacklistCallback(const FunctionCallbackInfo<Value> &info) {
    Isolate *isolate = info.GetIsolate();
    Local<Object> thisArg = info.Holder();
    Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

    if (info.Length() != 2) {
        isolate->ThrowException(String::NewFromUtf8(isolate, "blacklist requires 2 arguments"));
        return;
    }

    Local<Value> option = info[0];
    if (!option->IsString()) {
        isolate->ThrowException(String::NewFromUtf8(isolate, "option must be string"));
        return;
    }

    Local<Value> blackAddr = info[1];
    if (!blackAddr->IsString()) {
        isolate->ThrowException(String::NewFromUtf8(isolate, "blackAddr must be string"));
        return;
    }

    bool ret = fBlacklist(handler->Value(),*String::Utf8Value(isolate,option), *String::Utf8Value(isolate, blackAddr));
    info.GetReturnValue().Set(ret);
}
