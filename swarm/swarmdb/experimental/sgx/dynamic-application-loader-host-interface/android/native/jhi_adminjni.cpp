/*
   Copyright 2010-2016 Intel Corporation

   This software is licensed to you in accordance
   with the agreement between you and Intel Corporation.

   Alternatively, you can use this file in compliance
   with the Apache license, Version 2.


   Apache License, Version 2.0

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

/**                                                                            
********************************************************************************
**
**    @file jhi_adminjni.cpp
**
**    @brief  Defines and registers native interfaces of libjhi.so
**
**    @author Natalia Ovsyanikov
**
********************************************************************************
*/

#define LOG_TAG "JHI_JNI"

#include <map>
#include <stdio.h>
#include <string.h>
#include <stdio.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <sys/system_properties.h>

#include "misc.h"
#include "string_s.h"
#include "utils/Log.h"
#include "jni.h"
#include "JNIHelp.h"
#include "jhi.h"
#include "dbg.h"
#include "android_runtime/AndroidRuntime.h"

static JavaVM *gJavaVM;

#ifdef __cplusplus
extern "C" {
#endif  

static struct {
	void *handle = 0;
	bool isInitialized = false;
} gServiceHandle;

jobject gDalInfojObj;

char app_tmp_name[PROP_VALUE_MAX];

static void VerifyJhiHandler()
{
	char app_repo[PROP_VALUE_MAX];
	int ret;

	if (gServiceHandle.isInitialized) 
		return;

	memset(app_tmp_name, 0, sizeof(app_tmp_name));
	JHI_RET rc = JHI_Initialize((&gServiceHandle.handle), 0, 0);

	if (rc==JHI_SUCCESS)
		gServiceHandle.isInitialized = true;

	ret = __system_property_get("persist.jhi.APPLETS_LOCALE", app_repo);
	if (0 == ret)
		strcpy(app_repo, "/data/intel/dal/applet_repository");
	sprintf (app_tmp_name, "%s/tmp-", app_repo);
}

JNIEXPORT int JNICALL Java_com_intel_security_dalservice_JNIDALAdmin_DAL_Install_FD(
	JNIEnv *env, jclass cls,
	jstring AppId,
	jint AppFd,
	jint AppSize)
{
	JHI_RET ret = JHI_SUCCESS;
	FILE *pFile = NULL;
	void *mem = NULL;
	VerifyJhiHandler();
	string tmpApplet(app_tmp_name);
	const char *pAppId = env->GetStringUTFChars(AppId, NULL);
	TRACE0("JHI_JNI_ADMIN: Install_FD DAL Applet to JHI....");
	if (pAppId == NULL || AppFd == 0 || AppSize == 0 ){
		TRACE0("JHI_JNI_ADMIN: Install_FD Applet Failure");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}  

	if (!gServiceHandle.isInitialized) {
		TRACE0("JHI_JNI_ADMIN: Install_FD Applet Failure. init handle failed");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}

	tmpApplet += const_cast< char*>(pAppId);
	tmpApplet += ".dalp";
	TRACE2("JHI_JNI_ADMIN:DAL_Install_FD: tmpApplet %s, length %d \n",
		tmpApplet.c_str(), AppSize);
	pFile = fopen(reinterpret_cast<const char *>(tmpApplet.c_str()), "wb");
	mem = mmap(0, AppSize, PROT_READ, MAP_SHARED, AppFd, 0) ;
	if (mem == NULL || pFile == NULL) {
		TRACE0("JHI_JNI_ADMIN: DAL_Install_FD mmap, fopen Failure");
		goto exit;
	}
 
	if (fwrite(mem, 1, AppSize, pFile) != AppSize) {
		TRACE0("JHI_JNI_ADMIN: DAL_Install_FD fwrite Failure");
		goto exit;
	}

	fclose(pFile);
	ret = JHI_Install2(gServiceHandle.handle, pAppId, reinterpret_cast<const char *>(tmpApplet.c_str()));
	if (ret == JHI_SUCCESS) {
		TRACE0("JHI_JNI_ADMIN: Install_FD Applet Success");
	} else {
		TRACE0("JHI_JNI_ADMIN: Install_FD Applet Failure");
	}
exit:
	if (pFile != NULL)
		remove(reinterpret_cast<const char *>(tmpApplet.c_str()));
	if (mem != NULL)
		munmap(mem, AppSize);
	if (AppFd > 0)
		close(AppFd);
	env->ReleaseStringUTFChars(AppId, pAppId);
	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALAdmin_DAL_Install(
	JNIEnv *env, jclass cls, 
	jstring AppId, jstring AppPath) {

	JHI_RET ret = JHI_SUCCESS;
	const char* pAppId = env->GetStringUTFChars(AppId, NULL);
	const char* pAppPath = env->GetStringUTFChars(AppPath, NULL);

	TRACE0("JHI_JNI_ADMIN: Install DAL Applet to JHI....");
	if (pAppId == NULL || pAppPath == NULL) {
		TRACE0("JHI_JNI_ADMIN: Install Applet Failure");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}

	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized) {
		TRACE0("JHI_JNI_ADMIN: Install Applet Failure. init handle failed");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}

	ret = JHI_Install2(gServiceHandle.handle, pAppId, pAppPath); 
	if (ret == JHI_SUCCESS) {
		TRACE0("JHI_JNI_ADMIN: Install Applet Success");
	} else {
		TRACE0("JHI_JNI_ADMIN: Install Applet Failure");
	}
exit:
	env->ReleaseStringUTFChars(AppId, pAppId);
	env->ReleaseStringUTFChars(AppPath, pAppPath);

	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALAdmin_DAL_Uninstall(
	JNIEnv *env, jclass cls, 
	jstring AppId)
{
	JHI_RET ret = JHI_SUCCESS;
	const char *pAppId = env->GetStringUTFChars(AppId, NULL);

	TRACE0("JHI_JNI_ADMIN: Uninstall DAL Applet to JHI....");
	if (pAppId == NULL) {
		TRACE0("JHI_JNI_ADMIN: Uninstall Applet Failure");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}
   
	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized) {
		TRACE0("JHI_JNI_ADMIN: Uninstall Applet Failure. init handle failed");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}

	ret = JHI_Uninstall(gServiceHandle.handle, 
		const_cast<char *>(pAppId)); 
	if (ret == JHI_SUCCESS) {
		TRACE0("JHI_JNI_ADMIN: Uninstall Applet Success");
	} else {
		TRACE0("JHI_JNI_ADMIN: Uninstall Applet Failure");
	}
exit:
	env->ReleaseStringUTFChars(AppId, pAppId);

	return ret;
}

JNIEXPORT jobject JNICALL Java_com_intel_security_dalservice_JNIDALAdmin_DAL_GetVersionInfo(
	JNIEnv *env, jclass cls, jintArray retcode) {

	JHI_VERSION_INFO info;
	JHI_RET ret;
	jobject jObjDataVersionInfo = NULL;
	jmethodID constrDataVersionInfo = NULL;
	jclass jClassDataVersionInfo = NULL;
	jstring jhiVersion = NULL;
	jstring fwVersion = NULL;

	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized) {
		TRACE0("JHI_JNI_ADMIN: GetVersionInfo Failure. init handle failed");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}

	ret = JHI_GetVersionInfo(gServiceHandle.handle, &info);
	if (ret != JHI_SUCCESS)
		goto exit;

	jhiVersion = env->NewStringUTF(info.jhi_version);
	fwVersion = env->NewStringUTF(info.fw_version);

	jClassDataVersionInfo = env->GetObjectClass(gDalInfojObj);
	if (!jClassDataVersionInfo) {
		TRACE1("JHI_JNI_ADMIN:GetVersionInfo: Failed to get %s jclass",
			"com/intel/security/dalinterface/DALVersionInfo");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}

	constrDataVersionInfo = env->GetMethodID(jClassDataVersionInfo,
		"<init>",
		"(Ljava/lang/String;Ljava/lang/String;II)V");
	if (!constrDataVersionInfo) {
		TRACE1("JHI_JNI_ADMIN:GetVersionInfo: Failed to get %s constructor",
			"com/intel/security/dalinterface/DALVersionInfo");
		ret = JHI_INTERNAL_ERROR;
		goto exit;
	}

	jObjDataVersionInfo = env->NewObject(jClassDataVersionInfo, 
		constrDataVersionInfo,
		jhiVersion, fwVersion,
		info.comm_type,
		info.platform_id);
	if (!jObjDataVersionInfo) {
		TRACE1("JHI_JNI_ADMIN:GetVersionInfo: Failed to create %s jobject",
			"com/intel/security/dalinterface/DALVersionInfo");
		ret = JHI_INTERNAL_ERROR;
	} 
exit:	
	env->SetIntArrayRegion(retcode, 0, 1, reinterpret_cast <const int*>(&ret));

	return jObjDataVersionInfo;
}

JNINativeMethod gJHIAdminMethods [] = {
	{
		"DAL_Install",
		"(Ljava/lang/String;Ljava/lang/String;)I",
		(void *) Java_com_intel_security_dalservice_JNIDALAdmin_DAL_Install
	},

	{
		"DAL_Install_FD",
		"(Ljava/lang/String;II)I",
		(void *) Java_com_intel_security_dalservice_JNIDALAdmin_DAL_Install_FD
	},

	{
		"DAL_Uninstall",
		"(Ljava/lang/String;)I",
		(void *) Java_com_intel_security_dalservice_JNIDALAdmin_DAL_Uninstall
	},

	{
		"DAL_GetVersionInfo",
		"([I)Lcom/intel/security/dalinterface/DALVersionInfo;",
		(void *) Java_com_intel_security_dalservice_JNIDALAdmin_DAL_GetVersionInfo
	}
};

jint JNI_OnLoad(JavaVM* vm, void* reserved)
{
	JNIEnv* env = NULL;
	jint result = -1;
	jclass jClazz;

	gJavaVM = vm;

	if (vm->GetEnv((void**)&env, JNI_VERSION_1_4) != JNI_OK) {
		TRACE0("JHI_JNI_ADMIN: GetEnv failed!");
		return result;
	}

	jClazz = env->FindClass("com/intel/security/dalservice/JNIDALAdmin");

	if (!jClazz) {
		TRACE1("OnLoad:FindClass %s failed!", "com/intel/security/dalservice/JNIDALAdmin");
		return result;
	}

	if (env->RegisterNatives(jClazz, gJHIAdminMethods, NELEM(gJHIAdminMethods)) != JNI_OK) {
		TRACE0 ("OnLoad:Failed to register native methods");
		return -1;
	}

	jclass jclassDalInfo = env->FindClass("com/intel/security/dalinterface/DALVersionInfo");

	if (!jclassDalInfo) {
		TRACE1("OnLoad:Failed to get %s jclass", "com/intel/security/dalinterface/DALVersionInfo");
	} else {
		jmethodID constr = env->GetMethodID(jclassDalInfo, "<init>", "()V");
		if (!constr) {
			TRACE1("OnLoad:Failed to get %s constractor", "com/intel/security/dalinterface/DALVersionInfo");
		} else {
			jobject jobjDalInfo = env->NewObject(jclassDalInfo, constr);
			if (!jobjDalInfo) {
				TRACE1("OnLoad:Failed to get %s jobject", "com/intel/security/dalinterface/DALVersionInfo");
			} else {
				gDalInfojObj = env->NewGlobalRef(jobjDalInfo);
			}
		}
	}
	return JNI_VERSION_1_4;
}


#ifdef __cplusplus
}
#endif
