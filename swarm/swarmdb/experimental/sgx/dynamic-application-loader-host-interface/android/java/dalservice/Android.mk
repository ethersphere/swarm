LOCAL_PATH := $(call my-dir)

#Create service
include $(CLEAR_VARS)

LOCAL_PACKAGE_NAME := DALServiceRunner

LOCAL_SRC_FILES := $(call all-java-files-under,src)

LOCAL_REQUIRED_MODULES := com.intel.security.dalinterface	

LOCAL_JAVA_LIBRARIES :=	\
	com.intel.security.dalinterface \
	framework

LOCAL_REQUIRED_MODULES := libadminjni_jhi libclientjni_jhi

LOCAL_SDK_VERSION := current

LOCAL_PROGUARD_ENABLED := disabled

LOCAL_CERTIFICATE := platform

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

include $(BUILD_PACKAGE)
