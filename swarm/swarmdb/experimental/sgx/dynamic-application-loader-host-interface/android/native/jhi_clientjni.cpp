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
**    @file jhi_clientjni.cpp
**
**    @brief  Defines and registers native interfaces of libjhi.so
**	      for DALTransport service	
**    @author Natalia Ovsyanikov
**
********************************************************************************
*/

#define LOG_TAG "JHI_JNI"

#include <string.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <sys/system_properties.h>
#include <map>
#include <fcntl.h>
#include <pthread.h>
#include <semaphore.h>

#include "jni.h"
#include "dbg.h"
#include "JNIHelp.h"
#include "android_runtime/AndroidRuntime.h"
#include "utils/Log.h"
#include "jhi.h"
#include "misc.h"
#include "string_s.h"

JavaVM *gDalJavaVM;
jobject gDalCallbackData;
jobject gDalCallback;

#ifdef __cplusplus
using namespace std;
extern "C" {
#endif  

static struct {

	void *handle = 0;
	bool isInitialized = false;

} gServiceHandle;

static sem_t gCallbackSemaphore;

static void SocketsCleanup() {

	char event_repo[PROP_VALUE_MAX];
	char cleanup_events [PROP_VALUE_MAX + 30];
	int ret = __system_property_get("persist.jhi.EVENT_LOCALE", event_repo);

	if (0 == ret)
		strcpy(event_repo, "/data/intel/dal/dynamic_sockets");
	sprintf(cleanup_events, "exec rm %s/*", event_repo);
	system(cleanup_events);
}

static void VerifyJhiHandler() {

	if (gServiceHandle.isInitialized)
		return;

	JHI_RET rc = JHI_Initialize((&gServiceHandle.handle), 0, 0);

	if (rc==JHI_SUCCESS) {
		gServiceHandle.isInitialized = true;
		if (sem_init(&gCallbackSemaphore, 0, 10) == (-1)) {
			TRACE1("JHI_CLIENT_JNI:Init of callback semaphore error: %s\n", strerror(errno));
		}
	}
}

struct _callback_data {
	jlong jSessionHandle;
	void *data;
	long dataLength;
	int dataType;
};

inline void *_allocCallbackData(JHI_SESSION_HANDLE SessionHandle,
				JHI_EVENT_DATA eventData) {

	struct _callback_data *_clData = (struct _callback_data *)malloc(sizeof(struct _callback_data));
	if (_clData == NULL)
		return NULL;

	_clData->data = malloc(eventData.datalen);
	if (_clData->data == NULL) {
		free (_clData);
		return NULL;
	}

	_clData->dataLength = eventData.datalen;
	_clData->dataType = eventData.dataType;
	_clData->jSessionHandle = reinterpret_cast <jlong>(SessionHandle);
	memcpy(_clData->data, eventData.data, eventData.datalen);
	return _clData;
}

inline void _freeCallbackData(void *clData) {

	struct _callback_data *_clData = (struct _callback_data *)clData;
	if (_clData == NULL)
		return;
	if (_clData->data == NULL)
		return;
	if (_clData->data != NULL) {
		free(_clData->data);
	}
	free (_clData);
}

void *callbackThread(void *args) {

	JNIEnv *env;
	jobject dalCallback = NULL;
	jclass jcCallback = NULL, jcService = NULL;
	jmethodID callbackConstructor = NULL, callback = NULL;
	jbyteArray dataArray = NULL;
	struct _callback_data *rData = (struct _callback_data *)args;

	if (args == NULL) {
		TRACE0("JHI_CLIENT_JNI:localCallback: invalid args\n");
		goto unlock;
	}

	if (gDalJavaVM->AttachCurrentThread(&env, NULL) < 0) {
		TRACE0 ("JHI_CLIENT_JNI:localCallback: failed to attach current thread\n");
		goto unlock;
	}

	jcCallback = env->GetObjectClass(gDalCallbackData);
	if (jcCallback == NULL) {
		TRACE0 ("JHI_CLIENT_JNI:localCallback: failed to get DALCallback class reference\n");
		goto out;
	}

	callbackConstructor = env->GetMethodID(jcCallback, "<init>", "(J[BB)V");

	if (!callbackConstructor) {
		TRACE0("JHI_CLIENT_JNI:localCallback: Failed to get constractor com/intel/security/dalinterface/IDALServiceCallbackListener");
		goto out;
	}

	dataArray = env->NewByteArray(rData->dataLength);
	if (!dataArray) {
		TRACE0("JHI_CLIENT_JNI:localCallback Failed to create java DataArray\n");
		goto out;
	}

	env->SetByteArrayRegion(dataArray, 0, rData->dataLength,
		reinterpret_cast<const jbyte *>(rData->data));

	dalCallback = env->NewObject(jcCallback,
		callbackConstructor,
		rData->jSessionHandle,
		dataArray,
		rData->dataType);
	if (!dalCallback) {
		TRACE0("JHI_CLIENT_JNI:localCallback Failed to create com/intel/security/dalinterface/DALVersionInfo");
		goto out;
	}

	jcService = env->GetObjectClass(gDalCallback);
	if (jcService == NULL) {
		TRACE0 ("JHI_CLIENT_JNI:localCallback: failed to get Listener class reference\n");
		goto out;
	}

	callback  = env->GetStaticMethodID(jcService, "DALcallbackHandler",
		"(Lcom/intel/security/dalinterface/DALCallback;)V");
	if (!callback) {
		TRACE0("JHI_CLIENT_JNI: localCallback:Failed to get callback method\n");
		goto out;
	}

	env->CallStaticVoidMethod(jcService, callback, dalCallback);
out:
	_freeCallbackData(args);
	gDalJavaVM->DetachCurrentThread();
unlock:
	if (sem_post(&gCallbackSemaphore) != 0)
		TRACE1("JHI_CLIENT_JNI:post of callback semaphore error: %s\n", strerror(errno));

	return NULL;
}

static void localCallback(JHI_SESSION_HANDLE SessionHandle,JHI_EVENT_DATA eventData) {

	struct _callback_data *_pthread_a = (struct _callback_data *)_allocCallbackData(SessionHandle, eventData);
	pthread_t _callbackThread;

	if (_pthread_a == NULL) {
		TRACE0("JHI_CLIENT_JNI:localCallback: can't allocate args\n");
		return;
	}

	if (sem_wait(&gCallbackSemaphore) != 0)
		TRACE1("JHI_CLIENT_JNI:wait of callback semaphore error: %s\n", strerror(errno));

	if (pthread_create(&_callbackThread, NULL, callbackThread, _pthread_a))
		TRACE0("JHI_CLIENT_JNI:localCallback:failed to create thread\n");
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALTransport_DAL_CreateSession(
	JNIEnv *env, jclass cls,
	jstring AppId, jint AppPid, jint flags,
	jbyteArray initBuffer,
	jlongArray SessionHandle) {

	const char* pAppId = env->GetStringUTFChars(AppId, NULL);
	void  * sessionHandler = NULL;
	jlong jSessionHandler = 0;
	DATA_BUFFER iBuff, *piBuff;

	if (initBuffer == NULL) {
		iBuff.buffer = NULL;
		iBuff.length = 0;
		TRACE0("JHI_CLIENT_JNI: CreateSession init buffer NULL");
	} else  {
		jboolean isCopy;
		iBuff.buffer = env->GetByteArrayElements (initBuffer, &isCopy);
		iBuff.length = env->GetArrayLength (initBuffer);
	}

	piBuff = &iBuff;

	if (iBuff.length == 0 || iBuff.buffer == 0) {
		TRACE0("JHI_CLIENT_JNI: CreateSession init buffer NULL");
		piBuff = NULL;
	} else  {
		piBuff = &iBuff;
	}

	if (!pAppId) {
		TRACE0("JHI_CLIENT_JNI: Can't receive AppId");
		return JHI_INTERNAL_ERROR;
	}

	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized)
		return JHI_INTERNAL_ERROR;

	JHI_RET ret = JHI_CreateSessionProcess(gServiceHandle.handle, pAppId,
		AppPid, flags, piBuff, &sessionHandler);

	env->ReleaseStringUTFChars(AppId, pAppId);
	if (initBuffer != NULL) {
		env->ReleaseByteArrayElements (initBuffer, (jbyte *)iBuff.buffer,
			JNI_ABORT);
	}

	if (ret == JHI_SUCCESS) {
		TRACE0("JHI_CLIENT_JNI: Create Session Success");
	}
	else {
		TRACE0("JHI_CLIENT_JNI: Create Session Failure");
		return ret;
	}
	jSessionHandler = reinterpret_cast<jlong>(sessionHandler);
	env->SetLongArrayRegion(SessionHandle, 0, 1, &jSessionHandler);

	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALTransport_DAL_CloseSession(
	JNIEnv *env, jclass cls, jlong SessionHandle) {

	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized)
		return JHI_INTERNAL_ERROR;
	void *pSessionHandle = reinterpret_cast<void *>(SessionHandle);

	JHI_RET ret = JHI_CloseSession(gServiceHandle.handle, &pSessionHandle);

	if (ret == JHI_SUCCESS)
		TRACE0("JHI_CLIENT_JNI: Close Session Success");
	else
		TRACE0("JHI_CLIENT_JNI: Close Session Failure");

	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALTransport_DAL_ClearSessions(
	JNIEnv *env, jclass cls, jint AppPid) {

	VerifyJhiHandler();

	if (!gServiceHandle.isInitialized)
		return JHI_INTERNAL_ERROR;

	JHI_RET ret = JHI_ClearSessions(gServiceHandle.handle, AppPid);

	if (ret == JHI_SUCCESS)
		TRACE1("JHI_CLIENT_JNI: ClearSessions Success pid %d", AppPid);
	else
		TRACE1("JHI_CLIENT_JNI: ClearSessions Failure pid %d", AppPid);

	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALTransport_DAL_SendAndReceive(
	JNIEnv *env, jclass cls, jlong SessionHandle, jint cmdId, jbyteArray tx,
	jbyteArray rx, jintArray rxn, jintArray res) {

	JVM_COMM_BUFFER commBuff;
	jboolean isCopy;

	commBuff.TxBuf->length = env->GetArrayLength(tx);
	commBuff.TxBuf->buffer = env->GetByteArrayElements(tx, &isCopy);
	commBuff.RxBuf->length = env->GetArrayLength(rx);
	commBuff.RxBuf->buffer = env->GetByteArrayElements(rx, &isCopy);

	jint *pRes = env->GetIntArrayElements(res, &isCopy);

	if (commBuff.TxBuf->length == 0 || commBuff.TxBuf->buffer == NULL) {
		TRACE0("JHI_CLIENT_JNI: Invalid commTx params\n");
		return JHI_INTERNAL_ERROR;
	}

	if (commBuff.RxBuf->length == 0 || commBuff.RxBuf->buffer == NULL) {
		TRACE0("JHI_CLIENT_JNI: Invalid commRx params\n");
		return JHI_INTERNAL_ERROR;
	}

	if (pRes == NULL) {
		TRACE0("JHI_CLIENT_JNI: Invalid res param\n");
		return JHI_INTERNAL_ERROR;
	}

	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized)
		return JHI_INTERNAL_ERROR;

	JHI_RET ret = JHI_SendAndRecv2(gServiceHandle.handle, (void *)SessionHandle,
		cmdId, &commBuff,pRes );

	if (ret != JHI_SUCCESS) {
		TRACE0 ("JHI_CLIENT_JNI: SendAndReceive failed");
	} else {
		TRACE0 ("JHI_CLIENT_JNI: SendAndReceive success");
		env->SetIntArrayRegion(rxn, 0, 1, (jint*) &commBuff.RxBuf->length);
	}

	env->ReleaseByteArrayElements (tx,(jbyte *)commBuff.TxBuf->buffer,  JNI_ABORT);
	env->ReleaseByteArrayElements (rx,(jbyte *)commBuff.RxBuf->buffer, 0);
	env->ReleaseIntArrayElements (res, pRes, 0);

	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALTransport_DAL_RegisterEvents(
	JNIEnv *env, jclass cls, jlong SessionHandle)
{
	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized)
		return JHI_INTERNAL_ERROR;

	JHI_RET ret = JHI_RegisterEvents(gServiceHandle.handle, (void *)SessionHandle, localCallback);

	if (ret == JHI_SUCCESS)
		TRACE0("JHI_CLIENT_JNI: RegisterEvents Success");
	else
		TRACE0("JHI_CLIENT_JNI: RegisterEvents Failure");

	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALTransport_DAL_UnregisterEvents(
	JNIEnv *env, jclass cls, jlong SessionHandle)
{

	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized)
		return JHI_INTERNAL_ERROR;

	JHI_RET ret = JHI_UnRegisterEvents(gServiceHandle.handle, (void *) SessionHandle);

	if (ret == JHI_SUCCESS)
		TRACE0("JHI_CLIENT_JNI: UnregisterEvents Success");
	else
		TRACE0("JHI_CLIENT_JNI: UnregisterEvents Failure");

	return ret;
}

JNIEXPORT int  JNICALL Java_com_intel_security_dalservice_JNIDALTransport_DAL_SHMemTxRxTrans(
	JNIEnv *env, jclass cls,
	jlong SessionHandle,
	jint nCommandId,
	jint rfd,
	jint txLength,
	jintArray rxLength,
	jintArray responseCode) {

	unsigned char *mem = NULL;
	JVM_COMM_BUFFER commBuff;
	jboolean isCopy;
	JHI_RET ret = JHI_INTERNAL_ERROR;
	int Res = 0, *pRes = NULL;

	if (rxLength == NULL) {
		commBuff.RxBuf[0].length = 0;
		commBuff.RxBuf[0].buffer = NULL;

	} else {
		env->GetIntArrayRegion (rxLength, 0, 1, (jint *)&(commBuff.RxBuf[0].length));
	}

	if (rfd == 0 && (txLength != 0 || (rxLength != NULL && commBuff.RxBuf[0].length != 0))) {
		return JHI_INVALID_COMM_BUFFER;
	}

	commBuff.TxBuf[0].length = txLength;
	if (txLength != 0) {
		commBuff.TxBuf[0].buffer = JHI_ALLOC (commBuff.TxBuf[0].length);
		if (commBuff.TxBuf->buffer == NULL) {
			TRACE0("JHI_CLIENT_JNI: Can't allocate memory\n");
			return ret;
		}
	} else {
		commBuff.TxBuf[0].buffer = NULL;
	}

	if (commBuff.RxBuf[0].length != 0) {
		commBuff.RxBuf[0].buffer = JHI_ALLOC(commBuff.RxBuf[0].length);
		if (commBuff.RxBuf->buffer == NULL) {
			TRACE0("JHI_CLIENT_JNI: Can't Allocate Memory\n");
			goto exit;
		}
	} else {
		commBuff.RxBuf[0].buffer = NULL;
	}

	TRACE2("jHI_CLIENT_JNI: before DAL_SHMemTxRxTrans tx data l %d, rx data l %d\n",
		commBuff.TxBuf->length, commBuff.RxBuf->length);
	if (responseCode != NULL) {
		pRes = &Res;
	}

	VerifyJhiHandler();
	if (!gServiceHandle.isInitialized) {
		goto exit;
	}

	if (rfd != 0 && ((commBuff.TxBuf->length + commBuff.RxBuf->length) != 0)) {
		mem = (unsigned char *) mmap(0, commBuff.TxBuf->length + commBuff.RxBuf->length,
			PROT_READ|PROT_WRITE,
			MAP_SHARED, rfd, 0);
		if (mem == NULL) {
			TRACE0("JHI_CLIENT_JNI: DAL_SHMemTxRxTrans mmap Failure");
			goto exit;
		}

		ZeroMemory(commBuff.TxBuf->buffer, commBuff.TxBuf->length);
		ZeroMemory(commBuff.RxBuf->buffer, commBuff.RxBuf->length);
		memcpy(commBuff.TxBuf->buffer, mem, commBuff.TxBuf->length);
	}

	ret = JHI_SendAndRecv2(gServiceHandle.handle, (void *)SessionHandle,
		nCommandId, &commBuff,pRes );
	if (ret != JHI_SUCCESS) {
		TRACE0 ("JHI_CLIENT_JNI: SendAndReceive failed");
		goto exit;
	}
	TRACE0 ("JHI_CLIENT_JNI: SendAndReceive success");

	if (rxLength != 0)
		env->SetIntArrayRegion(rxLength, 0, 1, (jint *)&commBuff.RxBuf->length);
	if (mem != 0 && commBuff.RxBuf->length != 0) {
		TRACE2 ("jHI_CLIENT_JNI: after DAL_SHMemTxRxTrans tx data l %d, rx data l %d\n",
			commBuff.TxBuf->length, commBuff.RxBuf->length);
		memcpy (mem + commBuff.TxBuf->length, commBuff.RxBuf->buffer,
			commBuff.RxBuf->length);
	}

	if (pRes != NULL) {
		env->SetIntArrayRegion(responseCode, 0, 1, (jint *)pRes);
	}
exit:
	if (commBuff.TxBuf->buffer)
		JHI_DEALLOC (commBuff.TxBuf->buffer);
	if (commBuff.RxBuf->buffer)
		JHI_DEALLOC (commBuff.RxBuf->buffer);
	if (mem)
		munmap (mem, commBuff.TxBuf->length + commBuff.RxBuf->length);
	if (rfd > 0)
		close(rfd);

	return ret;
}


JNINativeMethod gJHIClientMethods [] = {
	{
		"DAL_CreateSession",
		"(Ljava/lang/String;II[B[J)I",
		(void *) Java_com_intel_security_dalservice_JNIDALTransport_DAL_CreateSession
	},

	{
		"DAL_CloseSession",
		"(J)I",
		(void *) Java_com_intel_security_dalservice_JNIDALTransport_DAL_CloseSession
	},

	{
		"DAL_SendAndRecv",
		"(JI[B[B[I[I)I",
		(void *) Java_com_intel_security_dalservice_JNIDALTransport_DAL_SendAndReceive
	},

	{
		"DAL_RegisterEvents",
		"(J)I",
		(void *) Java_com_intel_security_dalservice_JNIDALTransport_DAL_RegisterEvents
	},

	{
		"DAL_UnregisterEvents",
		"(J)I",
		(void *) Java_com_intel_security_dalservice_JNIDALTransport_DAL_UnregisterEvents
	},

	{
		"DAL_SHMemTxRxTrans",
		"(JIII[I[I)I",
		(void *) Java_com_intel_security_dalservice_JNIDALTransport_DAL_SHMemTxRxTrans
	},

	{
		"DAL_ClearSessions",
		"(I)I",
		(void *) Java_com_intel_security_dalservice_JNIDALTransport_DAL_ClearSessions
	}
};

jint JNI_OnLoad(JavaVM* vm, void* reserved)
{
	JNIEnv* env = NULL;
	jint result = -1;
	jclass jClazz;

	if (vm->GetEnv((void**)&env, JNI_VERSION_1_4) != JNI_OK) {
		TRACE0("JHI_CLIENT_JNI: GetEnv failed!");
		return result;
	}

	gDalJavaVM = vm;

	jClazz = env->FindClass("com/intel/security/dalservice/JNIDALTransport");

	if (!jClazz) {
		TRACE0("JHI_CLIENT_JNI: FindClass failed!");
		return result;
	}  

	if (env->RegisterNatives(jClazz, gJHIClientMethods, NELEM(gJHIClientMethods)) != JNI_OK) {
		TRACE0 ("JHI_CLIENT_JNI: Failed to register native methods");
		return -1;
	} 

	jclass jclassDalCallback = env->FindClass("com/intel/security/dalinterface/DALCallback");

	if (!jclassDalCallback) {
		TRACE0("JHI_CLIENT_JNI: OnLoad:Failed to get jclass com/intel/security/dalinterface/DALCallback");
	} else {
		jmethodID constr = env->GetMethodID(jclassDalCallback, "<init>", "(J[BB)V");

		if (!constr) {
			TRACE0("JHI_CLIENT_JNI: OnLoad:Failed to get constractor com/intel/security/dalinterface/DALCallback");
		} else {
			jbyteArray blankArray = env->NewByteArray(0);
			jobject jobjDalCallback = env->NewObject(jclassDalCallback, constr, (jlong)0, blankArray, (jbyte)0);

			if (!jobjDalCallback) {
				TRACE0("JHI_CLIENT_JNI: OnLoad:Failed to get jobject com/intel/security/dalinterface/DALCallback");
			} else {
				gDalCallbackData = env->NewGlobalRef(jobjDalCallback);
			}
		}
	}
	jclass jcService = env->FindClass("com/intel/security/dalservice/DALTransportServiceImpl");
	if (!jcService) {
		TRACE0("JHI_CLIENT_JNI: OnLoad:Failed to get jclass com/intel/security/dalservice/DALTransportServiceImpl\n");
		return JNI_VERSION_1_4;
	}
	jmethodID constrS = env->GetMethodID(jcService, "<init>", "()V");
	if (!constrS) {
		TRACE0("JHI_CLIENT_JNI: OnLoad:Failed to get service constructor\n");
		return JNI_VERSION_1_4;

	}
	jmethodID callback = env->GetStaticMethodID(jcService, "DALcallbackHandler",
		"(Lcom/intel/security/dalinterface/DALCallback;)V");
	if (!callback) {
		TRACE0("JHI_CLIENT_JNI: OnLoad:Failed to get callback method\n");
		return JNI_VERSION_1_4;
	}

	jobject joService = env->NewObject (jcService, constrS, NULL);
	if (!joService) {
		TRACE0("JHI_CLIENT_JNI: OnLoad:Failed to create service object\n");
		return JNI_VERSION_1_4;
	}

	gDalCallback = env->NewGlobalRef(joService);

	return JNI_VERSION_1_4;
}

#ifdef __cplusplus
}
#endif
