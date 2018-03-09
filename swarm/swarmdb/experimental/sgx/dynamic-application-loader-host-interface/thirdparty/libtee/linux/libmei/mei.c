/*
Intel Management Engine Interface (Intel MEI) Linux driver
Intel MEI Interface

This file is provided under BSD license.

BSD LICENSE

Copyright (c) 2003 - 2017 Intel Corporation.
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name Intel Corporation nor the
      names of its contributors may be used to endorse or promote products
      derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
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

#include "libmei.h"

/*****************************************************************************
 * Intel Management Engine Interface
 *****************************************************************************/
#ifdef __ANDROID__
#define LOG_TAG "libmei"
#include <cutils/log.h>
#define mei_msg(_me, fmt, ARGS...) ALOGV_IF(_me->verbose, fmt, ##ARGS)
#define mei_err(_me, fmt, ARGS...) ALOGE(fmt, ##ARGS)
static inline void __dump_buffer(const char *buf)
{
	ALOGV("%s\n", buf);
}

#else /* ! ANDROID */
#define mei_msg(_me, fmt, ARGS...) do {         \
	if (_me->verbose)                       \
		fprintf(stderr, "me: " fmt, ##ARGS);	\
} while (0)

#define mei_err(_me, fmt, ARGS...) do {         \
	fprintf(stderr, "me: error: " fmt, ##ARGS); \
} while (0)
static inline void __dump_buffer(const char *buf)
{
	fprintf(stderr, "%s\n", buf);;
}
#endif /* ANDROID */

static void mei_dump_hex_buffer(const unsigned char *buf, size_t len)
{
	const size_t pbufsz = 16 * 3;
	char pbuf[pbufsz];
	int j = 0;
	while (len-- > 0) {
		snprintf(&pbuf[j], pbufsz - j, "%02X ", *buf++);
		j += 3;
		if (j == 16 * 3) {
			__dump_buffer(pbuf);
			j = 0;
		}
	}
	if (j)
		__dump_buffer(pbuf);
}

void mei_deinit(struct mei *me)
{
	if (!me)
		return;

	if (me->fd != -1)
		close(me->fd);
	me->fd = -1;
	me->buf_size = 0;
	me->prot_ver = 0;
	me->state = MEI_CL_STATE_ZERO;
	me->last_err = 0;
}

static inline int __mei_errno_to_state(struct mei *me)
{
	switch(me->last_err) {
	case 0:         return me->state;
	case ENOTTY:    return MEI_CL_STATE_NOT_PRESENT;
	case EBUSY:     return MEI_CL_STATE_DISCONNECTED;
	case ENODEV:    return MEI_CL_STATE_DISCONNECTED;
	default:        return MEI_CL_STATE_ERROR;
	}
}

int mei_get_fd(struct mei *me)
{
	if (!me)
		return -EINVAL;
	return me->fd;
}

static inline int __mei_open(struct mei *me, const char *devname)
{
	errno = 0;
	me->fd = open(devname, O_RDWR);
	me->last_err = errno;
	return me->fd == -1 ? -me->last_err : me->fd;
}

static inline int __mei_connect(struct mei *me, struct mei_connect_client_data *d)
{
	errno = 0;
	int rc = ioctl(me->fd, IOCTL_MEI_CONNECT_CLIENT, d);
	me->last_err = errno;
	return rc == -1 ? -me->last_err : 0;
}

static inline ssize_t __mei_read(struct mei *me, unsigned char *buf, size_t len)
{
	ssize_t rc;
	errno = 0;
	rc = read(me->fd, buf, len);
	me->last_err = errno;
	return rc <= 0 ? -me->last_err : rc;
}

static inline ssize_t __mei_write(struct mei *me, const unsigned char *buf, size_t len)
{
	ssize_t rc;
	errno = 0;
	rc = write(me->fd, buf, len);
	me->last_err = errno;
	return rc <= 0 ? -me->last_err : rc;
}

int mei_init(struct mei *me, const char *device, const uuid_le *guid,
		unsigned char req_protocol_version, bool verbose)
{
	int rc;

	if (!me || !device || !guid)
		return -EINVAL;

	/* if me is uninitialized it will close wrong file descriptor */
	me->fd = -1;
	mei_deinit(me);

	me->verbose = verbose;

	mei_msg(me, "API version %hhd.%hhd\n",
		mei_get_api_version() >> 16, mei_get_api_version() >> 8);

	rc = __mei_open(me, device);
	if (rc < 0) {
		mei_err(me, "Cannot establish a handle to the Intel MEI driver %.20s [%d]:%s\n",
			device, rc, strerror(-rc));
		return rc;
	}

	mei_msg(me, "Opened %.20s: fd = %d\n", device, me->fd);

	memcpy(&me->guid, guid, sizeof(*guid));
	me->prot_ver = req_protocol_version;

	me->state = MEI_CL_STATE_INTIALIZED;

	return 0;
}

struct mei *mei_alloc(const char *device, const uuid_le *guid,
		unsigned char req_protocol_version, bool verbose)
{
	struct mei *me;

	if (!device || !guid)
		return NULL;

	me = malloc(sizeof(struct mei));
	if (!me)
		return NULL;

	if (mei_init(me, device, guid, req_protocol_version, verbose)) {
		free(me);
		return NULL;
	}
	return me;
}

void mei_free(struct mei *me)
{
	if (!me)
		return;
	mei_deinit(me);
	free(me);
}

int mei_connect(struct mei *me)
{
	struct mei_client *cl;
	struct mei_connect_client_data data;
	int rc;

	if (!me)
		return -EINVAL;

	if (me->state != MEI_CL_STATE_INTIALIZED &&
	    me->state != MEI_CL_STATE_DISCONNECTED) {
		mei_err(me, "client state [%d]\n", me->state);
		return -EINVAL;
	}

	memset(&data, 0, sizeof(data));
	memcpy(&data.in_client_uuid, &me->guid, sizeof(me->guid));

	rc = __mei_connect(me, &data);
	if (rc < 0) {
		me->state = __mei_errno_to_state(me);
		mei_err(me, "Cannot connect to client [%d]:%s\n", rc, strerror(-rc));
		return rc;
	}

	cl = &data.out_client_properties;
	mei_msg(me, "max_message_length %d\n", cl->max_msg_length);
	mei_msg(me, "protocol_version %d\n", cl->protocol_version);

	if ((me->prot_ver > 0) && (cl->protocol_version < me->prot_ver)) {
		mei_err(me, "Intel MEI protocol version not supported\n");
		me->state =  MEI_CL_STATE_VERSION_MISMATCH;
		rc = -EINVAL;
	} else {
		me->buf_size = cl->max_msg_length;
		me->prot_ver = cl->protocol_version;
		me->state =  MEI_CL_STATE_CONNECTED;
	}

	return rc ;
}

ssize_t mei_recv_msg(struct mei *me, unsigned char *buffer, size_t len)
{
	ssize_t rc;

	if (!me || !buffer)
		return -EINVAL;

	mei_msg(me, "call read length = %zd\n", len);

	rc = __mei_read(me, buffer, len);
	if (rc < 0) {
		me->state = __mei_errno_to_state(me);
		mei_err(me, "read failed with status [%zd]:%s\n", rc, strerror(-rc));
		goto out;
	}
	mei_msg(me, "read succeeded with result %zd\n", rc);
	if (me->verbose)
		mei_dump_hex_buffer(buffer, rc);
out:
	return rc;
}

ssize_t mei_send_msg(struct mei *me, const unsigned char *buffer, size_t len)
{
	ssize_t rc;

	if (!me || !buffer)
		return -EINVAL;

	mei_msg(me, "call write length = %zd\n", len);
	if (me->verbose)
		mei_dump_hex_buffer(buffer, len);

	rc  = __mei_write(me, buffer, len);
	if (rc < 0) {
		me->state = __mei_errno_to_state(me);
		mei_err(me, "write failed with status [%zd]:%s\n",
			rc, strerror(-rc));
		return rc;
	}

	return rc;
}

unsigned int mei_get_api_version()
{
	return LIBMEI_API_VERSION;
}
