LOCAL_PATH := $(call my-dir)

# Build the library
include $(CLEAR_VARS)

LOCAL_MODULE := com.intel.security.dalinterface

LOCAL_SRC_FILES := $(call all-java-files-under, .)
LOCAL_SRC_FILES += \
    com/intel/security/dalinterface/IDALTransportManager.aidl \
    com/intel/security/dalinterface/IDALAdminManager.aidl \
    com/intel/security/dalinterface/IDALServiceCallbackListener.aidl

LOCAL_MODULE_PATH := $(TARGET_OUT_JAVA_LIBRARIES)

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

include $(BUILD_JAVA_LIBRARY)

#Build the documentation
include $(CLEAR_VARS)
LOCAL_SRC_FILES := $(call all-subdir-java-files) $(call all-subdir-html-files)
LOCAL_MODULE := com.intel.security.dalinterface_doc
LOCAL_DROIDDOC_OPTIONS := com.intel.security.dalinterface
LOCAL_MODULE_CLASS := JAVA_LIBRARIES
LOCAL_DROIDDOC_USE_STANDARD_DOCLET := true
include $(BUILD_DROIDDOC)

# Copy com.intel.security.dalinterface.xml to /system/etc/permissions/
include $(CLEAR_VARS)
LOCAL_MODULE_TAGS := optional
LOCAL_MODULE := com.intel.security.dalinterface.xml
LOCAL_MODULE_CLASS := ETC
LOCAL_MODULE_PATH := $(TARGET_OUT_ETC)/permissions
LOCAL_SRC_FILES := $(LOCAL_MODULE)
include $(BUILD_PREBUILT)
