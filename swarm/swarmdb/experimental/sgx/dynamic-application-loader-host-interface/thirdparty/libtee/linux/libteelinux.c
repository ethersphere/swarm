/* Copyright 2014 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <unistd.h>
#include <errno.h>
#include <stdint.h>
#include <stdbool.h>
#include <linux/mei.h>
#include <helpers.h>
#include <libtee.h>
#include <libmei.h>

/* use inline function instead of macro to avoid -Waddress warning in GCC */
static inline struct mei *to_mei(PTEEHANDLE _h) __attribute__((always_inline));
static inline struct mei *to_mei(PTEEHANDLE _h)
{
	return _h ? (struct mei *)_h->handle : NULL;
}



static inline TEESTATUS errno2status(int err)
{
	switch (err) {
		case 0      : return TEE_SUCCESS;
		case -ENOTTY: return TEE_CLIENT_NOT_FOUND;
		case -EBUSY : return TEE_BUSY;
		case -ENODEV: return TEE_DISCONNECTED;
		default     : return TEE_INTERNAL_ERROR;
	}
}

TEESTATUS TEEAPI TeeInit(IN OUT PTEEHANDLE handle, IN const UUID *uuid, IN OPTIONAL const char *device)
{
	struct mei *me;
	TEESTATUS  status;

	FUNC_ENTRY();

	if (uuid == NULL || handle == NULL) {
		ERRPRINT("One of the parameters was illegal");
		status = TEE_INVALID_PARAMETER;
		goto End;
	}

	TEE_INIT_HANDLE(*handle);
	me = mei_alloc(device ? device : mei_default_device(), uuid, 0, false);
	if (!me) {
		ERRPRINT("Cannot init mei structure\n");
		status = TEE_INTERNAL_ERROR;
		goto End;
	}
	handle->handle = me;
	status = TEE_SUCCESS;

End:
	FUNC_EXIT(status);
	return status;
}

TEESTATUS TEEAPI TeeConnect(IN OUT PTEEHANDLE handle)
{
	struct mei *me = to_mei(handle);
	TEESTATUS  status;
	int        rc;


	FUNC_ENTRY();

	if (!me) {
		ERRPRINT("One of the parameters was illegal");
		status = TEE_INVALID_PARAMETER;
		goto End;
	}

	rc = mei_connect(me);
	if (rc) {
		ERRPRINT("Cannot establish a handle to the Intel MEI driver\n");
		status = errno2status(rc);
		goto End;
	}

	handle->maxMsgLen = me->buf_size;
	handle->protcolVer = me->prot_ver;

	status = TEE_SUCCESS;

End:
	FUNC_EXIT(status);
	return status;
}

TEESTATUS TEEAPI TeeRead(IN PTEEHANDLE handle, IN OUT void *buffer, IN size_t bufferSize,
			 OUT OPTIONAL size_t *pNumOfBytesRead)
{
	struct mei *me = to_mei(handle);
	TEESTATUS status;
	ssize_t rc;

	FUNC_ENTRY();

	if (!me || !buffer) {
		ERRPRINT("One of the parameters was illegal");
		status = TEE_INVALID_PARAMETER;
		goto End;
	}

	DBGPRINT("call read length = %zd\n", bufferSize);

	rc = mei_recv_msg(me, buffer, bufferSize);
	if (rc < 0) {
		status = errno2status(rc);
		ERRPRINT("read failed with status %zd %s\n",
				rc, strerror(rc));
		goto End;
	}

	status = TEE_SUCCESS;
	DBGPRINT("read succeeded with result %zd\n", rc);
	if (pNumOfBytesRead)
		*pNumOfBytesRead = rc;

End:
	FUNC_EXIT(status);
	return status;
}

TEESTATUS TEEAPI TeeWrite(IN PTEEHANDLE handle, IN const void *buffer, IN size_t bufferSize,
			  OUT OPTIONAL size_t *numberOfBytesWritten)
{
	struct mei *me  =  to_mei(handle);
	TEESTATUS status;
	ssize_t rc;

	FUNC_ENTRY();

	if (!me || !buffer) {
		ERRPRINT("One of the parameters was illegal");
		status = TEE_INVALID_PARAMETER;
		goto End;
	}

	DBGPRINT("call write length = %zd\n", bufferSize);

	rc  = mei_send_msg(me, buffer, bufferSize);
	if (rc < 0) {
		status = errno2status(rc);
		ERRPRINT("write failed with status %zd %s\n", rc, strerror(rc));
		goto End;
	}

	if (numberOfBytesWritten)
		*numberOfBytesWritten = rc;

	status = TEE_SUCCESS;
End:
	FUNC_EXIT(status);
	return status;
}

void TEEAPI TeeDisconnect(PTEEHANDLE handle)
{
	struct mei *me  =  to_mei(handle);
	FUNC_ENTRY();
	if (me) {
		mei_deinit(me);
		handle->handle = NULL;
	}

	FUNC_EXIT(TEE_SUCCESS);
}


TEESTATUS TEEAPI TeeCancel(IN PTEEHANDLE handle)
{
	FUNC_ENTRY();
	FUNC_EXIT(TEE_NOTSUPPORTED);
	return TEE_NOTSUPPORTED;
}

