LOCAL_PATH := $(call my-dir)

## TEE Library
include $(CLEAR_VARS)

LOCAL_MODULE_TAGS:= eng optional
LOCAL_SRC_FILES := linux/libteelinux.c linux/mei.c
LOCAL_MODULE    := libtee

LOCAL_COPY_HEADERS_TO := libtee
LOCAL_COPY_HEADERS := include/libtee/libtee.h
LOCAL_C_INCLUDES := $(LOCAL_PATH)/include
LOCAL_SHARED_LIBRARIES := libcutils

include $(BUILD_SHARED_LIBRARY)


ALL_DEFAULT_INSTALLED_MODULES += $(SYMLINKS)

# We need this so that the installed files could be picked up based on the
# local module name
ALL_MODULES.$(LOCAL_MODULE).INSTALLED := \
    $(ALL_MODULES.$(LOCAL_MODULE).INSTALLED) $(SYMLINKS)

