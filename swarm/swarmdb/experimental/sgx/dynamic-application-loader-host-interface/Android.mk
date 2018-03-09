TOP_PATH := $(call my-dir)
LOCAL_PATH := $(call my-dir)

####################
#   teetransport   #
####################
include $(CLEAR_VARS)

LOCAL_MODULE := libteetransport

LOCAL_SRC_FILES := \
    teetransport/teetransport.c \
    teetransport/teetransport_internal.c \
    teetransport/transport/libtee/teetransport_libtee.c \
    teetransport/transport/libtee/teetransport_libtee_wrapper.c \
    teetransport/transport/libtee/teetransport_libtee_client_metadata.c \
    thirdparty/libtee/linux/libteelinux.c \
    thirdparty/libtee/linux/libmei/mei.c

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/teetransport \
    $(LOCAL_PATH)/teetransport/transport/libtee \
    $(LOCAL_PATH)/teetransport/transport/socket \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/thirdparty/libtee/include/libtee \
    $(LOCAL_PATH)/thirdparty/libtee/linux/libmei

# Needed for libtee's prints
LOCAL_SHARED_LIBRARIES := liblog

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_SHARED_LIBRARY)

###################
#   bhplugin1.so  #
###################
include $(CLEAR_VARS)

LOCAL_MODULE := libbhplugin1

LOCAL_SRC_FILES := \
    thirdparty/bhplugin1/Beihai/tools/jhi_lib/BeihaiPlugin.cpp \
    plugins/bhplugin1/jhi_plugin.cpp \
    common/dbg-android.cpp \
    common/dbg.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/plugins/bhplugin1 \
    $(LOCAL_PATH)/thirdparty/bhplugin1/Beihai/tools/jhi_lib \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm \
    $(LOCAL_PATH)/common/include

LOCAL_SHARED_LIBRARIES := libteetransport liblog

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CPPFLAGS := -fexceptions
LOCAL_CFLAGS += -DDEBUG

include $(BUILD_SHARED_LIBRARY)

####################
#   bhplugin2.so   #
####################
include $(CLEAR_VARS)

LOCAL_MODULE := libbhplugin2

LOCAL_SRC_FILES := \
    /thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_impl.cpp \
    /thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_impl_admin.cpp \
    /thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_impl_ta.cpp \
    /thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_platform_linux.cpp \
    /thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/admin_pack/admin_pack_ext.c \
    /thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/admin_pack/admin_pack_int.c \
    /thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/admin_pack/bh_acp_util.c \
    plugins/bhplugin2/jhi_plugin.cpp \
    common/misc.cpp \
    common/dbg-android.cpp \
    common/dbg.cpp


LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/include \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/include \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm/beihai_shared/include \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm \
    $(LOCAL_PATH)/thirdparty/bhplugin2/include \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/plugins/bhplugin2

LOCAL_SHARED_LIBRARIES := libteetransport liblog

LOCAL_MODULE_TAGS:= optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CPPFLAGS := -fexceptions
LOCAL_CFLAGS += -DDEBUG

include $(BUILD_SHARED_LIBRARY)

##########################
#   SpoolerApplet.dalp   #
##########################
include $(CLEAR_VARS)

LOCAL_MODULE := SpoolerApplet.dalp
LOCAL_MODULE_CLASS := LIB
LOCAL_MODULE_PATH := $(TARGET_OUT)/vendor/intel/dal
LOCAL_SRC_FILES := applets/SpoolerApplet.dalp

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

include $(BUILD_PREBUILT)

#################
#   libjhi.so   #
#################
include $(CLEAR_VARS)

LOCAL_MODULE := libjhi

LOCAL_SRC_FILES := \
    libjhi/jhi.cpp \
    libjhi/CommandInvoker.cpp \
    libjhi/CommandsClientSocketsAndroid.cpp \
    common/dbg-android.cpp \
    common/dbg.cpp \
    common/locker-pthread.cpp \
    common/jhi_event_linux.cpp \
    common/reg-android.cpp \
    common/misc.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/libjhi \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm \
    $(LOCAL_PATH)/external

LOCAL_STATIC_LIBRARIES := jhi_uuid
LOCAL_SHARED_LIBRARIES := liblog

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_SHARED_LIBRARY)

#####################
#   teemanagement   #
#####################
include $(CLEAR_VARS)

LOCAL_MODULE := libteemanagement

LOCAL_SRC_FILES := \
    teemanagement/teemanagement.cpp \
    common/locker-pthread.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm \
    $(LOCAL_PATH)/libjhi

LOCAL_SHARED_LIBRARIES := libjhi

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_SHARED_LIBRARY)

############
#   jhid   #
############
include $(CLEAR_VARS)

LOCAL_MODULE := jhid

LOCAL_SRC_FILES := \
    service/AppletsManager.cpp \
    service/AppletsPackageReader.cpp \
    service/appProp.cpp \
    service/closeSession.cpp \
    service/CommandDispatcher.cpp \
    service/CommandsServerSocketsAndroid.cpp \
    service/createSession.cpp \
    service/DLL_Loader.cpp \
    service/EventManager.cpp \
    service/FWInfoLinux.cpp \
    service/getSCount.cpp \
    service/getSessionStat.cpp \
    service/GlobalsManager.cpp \
    service/init.cpp \
    service/install.cpp \
    service/JHIMain.cpp \
    service/jhi_plugin_loader.cpp \
    service/SendCmdPkg.cpp \
    service/LinuxService.cpp \
    service/ReadWriteLockPThread.cpp \
    service/sar.cpp \
    service/SessionsManager.cpp \
    service/uninstall.cpp \
    service/XmlReaderLibXml2.cpp \
    common/jhi_event_linux.cpp \
    common/jhi_semaphore-linux.cpp \
    common/misc.cpp \
    common/reg-android.cpp \
    common/locker-pthread.cpp \
    common/dbg-android.cpp \
    common/dbg.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/common/FWUpdate \
    $(LOCAL_PATH)/external/libxml2/include \
    $(LOCAL_PATH)/external/libb64-1.2/include \
    $(LOCAL_PATH)/external \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm

LOCAL_STATIC_LIBRARIES := jhi_uuid jhi_libxml2 libb64
LOCAL_SHARED_LIBRARIES := libdl liblog libteetransport

LOCAL_CPPFLAGS := -fexceptions

#$(shell (cd $(LOCAL_PATH)/jhid/; perl GenerateMessageFiles.pl))

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_EXECUTABLE)

#################
#   SmokeTest   #
#################
include $(CLEAR_VARS)

LOCAL_MODULE    := smoketest

LOCAL_SRC_FILES := test/smoketest/smoketest.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm \

LOCAL_CPPFLAGS := -fexceptions

LOCAL_STATIC_LIBRARIES := jhi_libxml2 libb64 jhi_uuid

LOCAL_SHARED_LIBRARIES := liblog libdl libjhi libteemanagement

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_EXECUTABLE)

############
#   BIST   #
############
include $(CLEAR_VARS)

LOCAL_MODULE := bist
LOCAL_SRC_FILES := test/bist/bist.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm

LOCAL_SHARED_LIBRARIES := libjhi libteemanagement

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_EXECUTABLE)

####################
#   adminjni_jhi   #
####################
include $(CLEAR_VARS)

LOCAL_MODULE := libadminjni_jhi

LOCAL_SRC_FILES := android/native/jhi_adminjni.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/external \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm \
    $(LOCAL_PATH)/thirdparty/bhplugin2/include

LOCAL_SHARED_LIBRARIES := libjhi liblog libcutils libnativehelper libandroid_runtime

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_SHARED_LIBRARY)

#####################
#   clientjni_jhi   #
#####################
include $(CLEAR_VARS)

LOCAL_MODULE := libclientjni_jhi

LOCAL_SRC_FILES := android/native/jhi_clientjni.cpp

LOCAL_C_INCLUDES := \
    $(LOCAL_PATH)/common/include \
    $(LOCAL_PATH)/external \
    $(LOCAL_PATH)/thirdparty/bhplugin2/FW/src/apps/dal_ivm \
    $(LOCAL_PATH)/thirdparty/bhplugin2/include              

LOCAL_SHARED_LIBRARIES := libjhi liblog libcutils libnativehelper libandroid_runtime

LOCAL_MODULE_TAGS := optional
LOCAL_PROPRIETARY_MODULE := true
LOCAL_MODULE_OWNER := intel

LOCAL_CFLAGS += -DDEBUG

include $(BUILD_SHARED_LIBRARY)

############
#   Java   #
############
SAVED_LOCAL_PATH := $(LOCAL_PATH)
include $(SAVED_LOCAL_PATH)/android/java/dalservice/Android.mk
include $(SAVED_LOCAL_PATH)/android/java/dalinterface/Android.mk

################
#   external   #
################
include $(SAVED_LOCAL_PATH)/external/uuid/Android.mk
include $(SAVED_LOCAL_PATH)/external/libxml2/Android.mk
include $(SAVED_LOCAL_PATH)/external/libb64-1.2/Android.mk

#$(LOCAL_PATH)/applets/Android.mk
