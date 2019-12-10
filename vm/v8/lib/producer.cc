#include "producer.h"
#include "../engine.h"
#include "memory.h"

static FuncProducer fproducer = NULL;

void InitializeProducer(FuncProducer producer) {
    fproducer = producer;
}

void NewProducerInstance(Isolate *isolate, Local<Context> context, void *address) {
    Local<ObjectTemplate> producerTpl = ObjectTemplate::New(isolate);
    producerTpl->SetInternalFieldCount(1);

    producerTpl->Set(String::NewFromUtf8(isolate, "producer"), FunctionTemplate::New(isolate, producerCallback),
                    static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));

    Local<Object> instance = producerTpl->NewInstance(context).ToLocalChecked();
    instance->SetInternalField(0, External::New(isolate, address));
    context->Global()->DefineOwnProperty(context, String::NewFromUtf8(isolate, "_native_producer"), instance,
                                         static_cast<PropertyAttribute>(PropertyAttribute::DontDelete | PropertyAttribute::ReadOnly));
}

// producerCallback
void producerCallback(const FunctionCallbackInfo<Value> &info) {
    Isolate *isolate = info.GetIsolate();
    Local<Object> thisArg = info.Holder();
    Local<External> handler = Local<External>::Cast(thisArg->GetInternalField(0));

    if (info.Length() != 2) {
        isolate->ThrowException(String::NewFromUtf8(isolate, "Producer requires 2 arguments"));
        return;
    }

    Local<Value> option = info[0];
    if (!option->IsString()) {
        isolate->ThrowException(String::NewFromUtf8(isolate, "option must be string"));
        return;
    }

    Local<Value> nodeAddr = info[1];
    if (!nodeAddr->IsString()) {
        isolate->ThrowException(String::NewFromUtf8(isolate, "nodeAddr must be string"));
        return;
    }

    bool ret = fproducer(handler->Value(),*String::Utf8Value(isolate,option), *String::Utf8Value(isolate, nodeAddr));
    info.GetReturnValue().Set(ret);
}
