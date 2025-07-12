#ifndef INTENT_H
#define INTENT_H

#include <stdlib.h>

struct Intent {
    const char* action;
    const char* type;
    const char* uri;
    const char* text;
};

void initJNI(uintptr_t javaVM, uintptr_t jniEnv, uintptr_t ctx);

struct Intent getIntent(uintptr_t javaVM, uintptr_t jniEnv, uintptr_t ctx);

void readContent(uintptr_t javaVM, uintptr_t jniEnv, uintptr_t ctx,
    const char* uri, uint8_t** output, uint32_t* outputLength);

#endif