LOCAL_PATH := $(call my-dir)
#b64
include $(CLEAR_VARS)

LOCAL_C_INCLUDES := $(LOCAL_PATH)/include
LOCAL_SRC_FILES :=  src/cdecode.c src/cencode.c
LOCAL_MODULE := libb64
LOCAL_MODULE_TAGS := optional

include $(BUILD_STATIC_LIBRARY)
