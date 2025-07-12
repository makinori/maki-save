// clang-format off
//go:build android
// clang-format on

#include "intent.h"

#include <jni.h>
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

// https://fynelabs.com/2024/03/01/running-native-android-code-in-a-fyne-app/
// https://github.com/fyne-io/fyne/blob/master/app/app_mobile_and.c
// https://github.com/fyne-io/fyne/blob/master/internal/driver/mobile/android.c
// https://github.com/fyne-io/fyne/blob/master/internal/driver/mobile/app/android.c

const char* jstringToC(JNIEnv* env, jstring str)
{
    const char* chars = (*env)->GetStringUTFChars(env, str, NULL);
    const char* copy = strdup(chars);
    (*env)->ReleaseStringUTFChars(env, str, chars);
    return copy;
}

// https://docs.oracle.com/javase/1.5.0/docs/guide/jni/spec/types.html#wp276
// https://docs.oracle.com/en/java/javase/17/docs/specs/jni/functions.html

// https://developer.android.com/reference/android/content/Context
static jmethodID contextGetContentResolverMethod;

// https://developer.android.com/reference/android/app/Activity
static jmethodID activityGetIntendMethod;

// https://developer.android.com/reference/android/content/Intent
static jmethodID intentGetActionMethod;
static jmethodID intentGetTypeMethod;
static jmethodID intentGetParcelableExtraMethod;
static jmethodID intentGetStringExtraMethod;

// https://developer.android.com/reference/android/net/Uri
static jmethodID uriParseMethod;
static jmethodID uriToStringMethod;

// https://developer.android.com/reference/android/content/ContentResolver
static jmethodID contentResolverOpenInputStreamMethod;

// https://docs.oracle.com/javase/8/docs/api/java/io/InputStream.html
static jmethodID inputStreamAvailableMethod;
static jmethodID inputStreamReadMethod;

void initJNI(uintptr_t javaVM, uintptr_t jniEnv, uintptr_t ctx)
{
    JNIEnv* env = (JNIEnv*)jniEnv;

    // cant store classes statically cause they're environment dependent?

    jclass contextClass = (*env)->FindClass(env, "android/content/Context");
    contextGetContentResolverMethod = (*env)->GetMethodID(env, contextClass,
        "getContentResolver", "()Landroid/content/ContentResolver;");

    jclass activityClass = (*env)->FindClass(env, "android/app/Activity");
    activityGetIntendMethod = (*env)->GetMethodID(env, activityClass,
        "getIntent", "()Landroid/content/Intent;");

    jclass intentClass = (*env)->FindClass(env, "android/content/Intent");
    intentGetActionMethod = (*env)->GetMethodID(env, intentClass,
        "getAction", "()Ljava/lang/String;");
    intentGetTypeMethod = (*env)->GetMethodID(env, intentClass,
        "getType", "()Ljava/lang/String;");
    intentGetParcelableExtraMethod = (*env)->GetMethodID(env, intentClass,
        "getParcelableExtra", "(Ljava/lang/String;)Landroid/os/Parcelable;");
    // api level 33 tiramisu
    // intentGetParcelableExtraMethod =
    //     (*env)->GetMethodID(env, intentClass, "getParcelableExtra",
    // "(Ljava/lang/String;Ljava/lang/Class;)Ljava/lang/Object;");
    intentGetStringExtraMethod = (*env)->GetMethodID(env, intentClass,
        "getStringExtra", "(Ljava/lang/String;)Ljava/lang/String;");

    jclass uriClass = (*env)->FindClass(env, "android/net/Uri");
    uriParseMethod = (*env)->GetStaticMethodID(env, uriClass,
        "parse", "(Ljava/lang/String;)Landroid/net/Uri;");
    uriToStringMethod = (*env)->GetMethodID(env, uriClass,
        "toString", "()Ljava/lang/String;");

    jclass contentResolverClass = (*env)->FindClass(env, "android/content/ContentResolver");
    contentResolverOpenInputStreamMethod = (*env)->GetMethodID(env, contentResolverClass,
        "openInputStream", "(Landroid/net/Uri;)Ljava/io/InputStream;");

    jclass inputStreamClass = (*env)->FindClass(env, "java/io/InputStream");
    inputStreamAvailableMethod = (*env)->GetMethodID(env, inputStreamClass,
        "available", "()I");
    inputStreamReadMethod = (*env)->GetMethodID(env, inputStreamClass,
        "read", "([B)I");

    // clean up with
    // (*env)->DeleteLocalRef(env, ...);
}

struct Intent getIntent(uintptr_t javaVM, uintptr_t jniEnv, uintptr_t ctx)
{
    JNIEnv* env = (JNIEnv*)jniEnv;

    struct Intent out = {};

    jobject intent = (*env)->CallObjectMethod(env, (jobject)ctx, activityGetIntendMethod);
    if (intent == NULL) {
        return out;
    }

    jstring action = (*env)->CallObjectMethod(env, intent, intentGetActionMethod);
    if (action != NULL) {
        out.action = jstringToC(env, action);
    }

    jstring type = (*env)->CallObjectMethod(env, intent, intentGetTypeMethod);
    if (type != NULL) {
        out.type = jstringToC(env, type);
    }

    // get uri

    jstring EXTRA_STREAM = (*env)->NewStringUTF(env, "android.intent.extra.STREAM");

    jobject uri = (*env)->CallObjectMethod(env, intent,
        intentGetParcelableExtraMethod, EXTRA_STREAM);

    if (uri != NULL) {
        jstring uriString = (*env)->CallObjectMethod(env, uri, uriToStringMethod);
        if (uriString != NULL) {
            out.uri = jstringToC(env, uriString);
        }
    }

    // get text

    jstring EXTRA_TEXT = (*env)->NewStringUTF(env, "android.intent.extra.TEXT");

    jstring text = (*env)->CallObjectMethod(env, intent,
        intentGetStringExtraMethod, EXTRA_TEXT);

    if (text != NULL) {
        out.text = jstringToC(env, text);
    }

    return out;
}

void readContent(
    uintptr_t javaVM, uintptr_t jniEnv, uintptr_t ctx,
    const char* uriString, uint8_t** output, uint32_t* outputLength)
{
    JNIEnv* env = (JNIEnv*)jniEnv;

    jstring uriJstring = (*env)->NewStringUTF(env, uriString);

    jclass uriClass = (*env)->FindClass(env, "android/net/Uri");
    jobject uri = (*env)->CallStaticObjectMethod(env, uriClass, uriParseMethod,
        uriJstring);

    jobject contentResolver = (*env)->CallObjectMethod(env, (jobject)ctx,
        contextGetContentResolverMethod);

    jobject inputStream = (*env)->CallObjectMethod(env, contentResolver,
        contentResolverOpenInputStreamMethod, uri);

    // considering it's a file, probably isn't chunked

    jint availableBytes = (*env)->CallIntMethod(env, inputStream,
        inputStreamAvailableMethod);

    jbyteArray byteArray = (*env)->NewByteArray(env, availableBytes);

    jint readBytes = (*env)->CallIntMethod(env, inputStream,
        inputStreamReadMethod, byteArray);

    // assert availableBytes == readBytes

    jbyte* dataPtr = (*env)->GetByteArrayElements(env, byteArray, JNI_FALSE); // false: direct pointer
    jsize dataLength = (*env)->GetArrayLength(env, byteArray);

    *output = malloc(dataLength);
    *outputLength = dataLength;
    memcpy(*output, dataPtr, dataLength);

    (*env)->ReleaseByteArrayElements(env, byteArray, dataPtr, JNI_ABORT);
}
