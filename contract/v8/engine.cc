// Copyright 2015 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <v8.h>
#include <libplatform/libplatform.h>
#include "engine.h"
#include "lib/blockchain.h"
#include "lib/load_lib.h"
#include "lib/load_sc.h"
#include "lib/storage.h"
#include "lib/reward_distributor.h"

using namespace v8;
std::unique_ptr<Platform> platformPtr;

void Initialize(){
    // Initialize V8.
    platformPtr = platform::NewDefaultPlatform();
    V8::InitializePlatform(platformPtr.get());
    V8::Initialize();
}

int executeV8Script(const char *sourceCode, uintptr_t handler, char **result) {

  // Create a new Isolate and make it the current one.
  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = ArrayBuffer::Allocator::NewDefaultAllocator();
  Isolate* isolate = Isolate::New(create_params);

  {
    Isolate::Scope isolate_scope(isolate);

    // Create a stack-allocated handle scope.
    HandleScope handle_scope(isolate);
    //
    Local<ObjectTemplate> globalTpl = NewNativeRequireFunction(isolate);
    // Create a new context.
    Local<Context> context = v8::Context::New(isolate, NULL, globalTpl);

    // Enter the context for compiling and running the hello world script.
    Context::Scope context_scope(context);

    NewBlockchainInstance(isolate, context, (void *)handler);
    NewStorageInstance(isolate, context, (void *)handler);
    NewRewardDistributorInstance(isolate, context, (void *)handler);

    LoadLibraries(isolate, context);
    {

      // Create a string containing the JavaScript source code.
      Local<String> source =
          String::NewFromUtf8(isolate, sourceCode,
                                  NewStringType::kNormal)
              .ToLocalChecked();

      // Compile the source code.
      Local<Script> script = Script::Compile(context, source).ToLocalChecked();

      // Run the script to get the result.
      Local<Value> scriptRes = script->Run(context).ToLocalChecked();

      // set result.
      if (result != NULL && !scriptRes.IsEmpty())  {
        Local<Object> obj = scriptRes.As<Object>();
        if (!obj->IsUndefined()) {
            String::Utf8Value str(isolate, obj);
            *result = (char *)malloc(str.length() + 1);
            strcpy(*result, *str);
        }
      }
    }
  }

  // Dispose the isolate and tear down V8.
  isolate->Dispose();

  delete create_params.array_buffer_allocator;
  return 0;
}

void DisposeV8(){
  V8::Dispose();
  V8::ShutdownPlatform();
}