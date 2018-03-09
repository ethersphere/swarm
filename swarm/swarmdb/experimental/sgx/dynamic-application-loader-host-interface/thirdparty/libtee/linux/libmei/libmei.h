/*
Intel Management Engine Interface (Intel MEI) Linux driver
Intel MEI Interface Header

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

/*! \file libmei.h
    \brief mei library API
 */

 #ifndef __LIBMEI_H__
#define __LIBMEI_H__

#include <linux/uuid.h>
#include <linux/mei.h>
#include <stdbool.h>
#include <stdint.h>
#include <unistd.h>

#ifdef __cplusplus
extern "C" {
#endif /*  __cplusplus */

/*! Library API version encode helper
 */
#define MEI_ENCODE_VERSION(major, minor)   ((major) << 16 | (minor) << 8)

/*! Library API version
 */
#define LIBMEI_API_VERSION MEI_ENCODE_VERSION(1, 0)

/*! Get current supported library API version
 *
 * \return version value
 */
unsigned int mei_get_api_version();

/*! ME client connection state
 */
enum mei_cl_state {
	MEI_CL_STATE_ZERO = 0,          /**< reserved */
	MEI_CL_STATE_INTIALIZED = 1,    /**< client is initialized */
	MEI_CL_STATE_CONNECTED,         /**< client is connected */
	MEI_CL_STATE_DISCONNECTED,      /**< client is disconnected */
	MEI_CL_STATE_NOT_PRESENT,       /**< client with GUID is not present in the system */
	MEI_CL_STATE_VERSION_MISMATCH,  /**< client version not supported */
	MEI_CL_STATE_ERROR,             /**< client is in error state */
};

/*! Structure to store connection data
 */
struct mei {
	uuid_le guid;           /**< client UUID */
	unsigned int buf_size;  /**< maximum buffer size supported by client*/
	unsigned char prot_ver; /**< protocol version */
	int fd;                 /**< connection file descriptor */
	int state;              /**< client connection state */
	int last_err;           /**< saved errno */
	bool verbose;           /**< verbose execution */
};

/*! find mei default device
 */
#ifndef ARRAY_SIZE
#define ARRAY_SIZE(a) (sizeof (a) / sizeof ((a)[0]))
#endif
static inline const char *mei_default_device()
{
	static const char *devnode[] = {"/dev/mei0", "/dev/mei"};
	int i;

	for (i = 0; i < ARRAY_SIZE(devnode); i++) {
		if (access(devnode[i], F_OK) == 0)
			return devnode[i];
	}
	return NULL;
}

/*! Allocate and initialize me handle structure
 *
 *  \param device device path, set MEI_DEFAULT_DEVICE to use default
 *  \param guid UUID/GUID of associated mei client
 *  \param req_protocol_version minimal required protocol version, 0 for any
 *  \param verbose print verbose output to console
 *  \return me handle to the mei device. All subsequent calls to the lib's functions
 *         must be with this handle. NULL on failure.
 */
struct mei *mei_alloc(const char *device, const uuid_le *guid,
		unsigned char req_protocol_version, bool verbose);

/*! Free me handle structure
 *
 *  \param me The mei handle
 */
void mei_free(struct mei *me);

/*! Initializes a mei connection
 *
 *  \param me A handle to the mei device. All subsequent calls to the lib's functions
 *         must be with this handle
 *  \param device device path, set MEI_DEFAULT_DEVICE to use default
 *  \param guid UUID/GUID of associated mei client
 *  \param req_protocol_version minimal required protocol version, 0 for any
 *  \param verbose print verbose output to a console
 *  \return 0 if successful, otherwise error code
 */
int mei_init(struct mei *me, const char *device, const uuid_le *guid,
		unsigned char req_protocol_version, bool verbose);

/*! Closes the session to me driver
 *  Make sure that you call this function as soon as you are done with the device,
 *  as other clients might be blocked until the session is closed.
 *
 *  \param me The mei handle
 */
void mei_deinit(struct mei *me);

/*! Open mei device and starts a session with an mei client
 *  If the application requested specific minimal protocol version
 *  and client doesn't support that version
 *  the handle state will be set to MEI_CL_STATE_VERSION_MISMATCH
 *  but connection will be established
 *
 *  \param me The mei handle
 *  \return 0 if successful, otherwise error code
 */
int mei_connect(struct mei *me);


/*! return file descriptor to opened handle
 *
 *  \param me The mei handle
 *  \return file descriptor or error
 */
int mei_get_fd(struct mei *me);

/*! Read data from the mei device.
 *
 *  \param me The mei handle
 *  \param buffer A pointer to a buffer that receives the data read from the mei device.
 *  \param len The number of bytes to be read.
 *  \return number of byte read if successful, otherwise error code
 */
ssize_t mei_recv_msg(struct mei *me, unsigned char *buffer, size_t len);

/*! Writes the specified buffer to the mei device.
 *
 *  \param me The mei handle
 *  \param buffer A pointer to the buffer containing the data to be written to the mei device.
 *  \param len The number of bytes to be written.
 *  \return number of bytes written if successful, otherwise error code
 */
ssize_t mei_send_msg(struct mei *me, const unsigned char *buffer, size_t len);

#ifdef __cplusplus
}
#endif /*  __cplusplus */

#endif /* __LIBMEI_H__ */
