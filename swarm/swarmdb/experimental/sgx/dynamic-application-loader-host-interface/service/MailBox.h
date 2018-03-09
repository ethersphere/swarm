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

/*#############################################################################
 # File: MailBox.h
 # Author: Yoel Gluschnaider.
 # Description:
 #    The mailbox is an inter process communication mechanism. It enables one
 #    way SMS's to be sent from process to process. For using the the MailBox
 #    module, see the MailBoxSendExample.c and MailBoxReadExample.c.
 #
 #############################################################################*/

#ifndef _MAIL_BOX_H_
#define _MAIL_BOX_H_

#include <windows.h>
#include <stdio.h>
#include <string.h>

// the maximal size of a MailSlot message.
#define MS_MAX_MESSAGE_SIZE 424
// The message key max size
#define MB_KEY_MAX_SIZE 16
// the message from and to fields and also the mailbox name max size.
#define MB_NAME_MAX_SIZE 32
// the maxinal message data size.
#define MB_DATA_MAX_SIZE 320


// return status of the mail box functions.
typedef enum _MB_RETURN_STATUS {
   MB_STATUS_OK,    
   MB_STATUS_FAIL_TO_OPEN, 
   MB_STATUS_FAIL_TO_CLOSE,
   MB_STATUS_SEND_ERR,
   MB_STATUS_READ_ERR,
   MB_STATUS_MB_NOT_OPEN,
   MB_STATUS_CHECK_ERR,
   MB_STATUS_ALREADY_OPEN,
   MB_STATUS_ILLEGAL_NAME,
   MB_STATUS_NO_PENDING_MESSAGES,
   MB_STATUS_INVALID_MODE,
   MB_STATUS_ILLEGAL_MSG_FORMAT,
   MB_STATUS_ILLEGAL_MSG_COUNT,
   MB_STATUS_ILLEGAL_PARAMS,
   MB_STATUS_NOT_OPEN,
   MB_STATUS_GENERAL_ERR
} MB_RETURN_STATUS;

// mode of the mailbox (read or write)
typedef enum _MB_MODE {
   MB_MODE_READ, 
   MB_MODE_WRITE
} MB_MODE;


// represent a mail box instance.
typedef struct {
   char name[MB_NAME_MAX_SIZE + 2];
   HANDLE handle;
   MB_MODE mode;
} MailBox;

// represent a message that can be sent/received using the mailbox.
typedef struct {
   char key[MB_KEY_MAX_SIZE + 2];
   char from[MB_NAME_MAX_SIZE + 2];
   char to[MB_NAME_MAX_SIZE + 2];
   char data[MB_DATA_MAX_SIZE + 2];
} MBMessage;


#ifdef __cplusplus
extern "C" {            /* Assume C declarations for C++ */
#endif	/* __cplusplus */



/*
 * Initializes the mail box struct.
 */
void MBMailBoxInit(MailBox *mailBox);

/*
 * Openes the specified mailbox in the given mode.
 * Returns:
 *    MB_STATUS_OK if the mailbox opened correctly
 *    MB_STATUS_ALREADY_OPEN if the mailbox was already open.
 *    MB_STATUS_ILLEGAL_NAME if name was longer than MB_NAME_MAX_SIZE or NULL.
 *    MB_ILLEGAL_MODE if the mode was not read or write.
 *    MB_STATUS_FAIL_TO_OPEN otherwise.
 */
MB_RETURN_STATUS MBOpenMailBox(MailBox *mailBox, const char *name, MB_MODE mode);

/*
 * Sends the message to the specified mailbox.
 * You must open the mailbox before and close it after the sending.
 * Returns:
 *    MB_STATUS_OK if the message was sent
 *    MB_STATUS_MB_NOT_OPEN if the mailbox isn't open
 *    MB_STATUS_INVALID_MODE if the mailbox is not in write mode.
 *    MB_STATUS_SEND_ERR otherwise.
 */
MB_RETURN_STATUS MBSendMessage(const MailBox *mailBox, const MBMessage &message);

/*
 * Sends the message to the specified mailbox. Opens and closes the mailbox automaticaly.
 * Returns:
 *    MB_STATUS_OK if the message was sent
 *    MB_STATUS_FAIL_TO_OPEN if the mailbox failed to open.
 *    MB_STATUS_SEND_ERR if the sending failed.
 *    MB_STATUS_ILLEGAL_NAME if name was longer than MB_NAME_MAX_SIZE or NULL.
 *    MB_STATUS_FAIL_TO_CLOSE if the mailbox failed to close.
 */
MB_RETURN_STATUS MBSendSingleMessage(const char *name, const MBMessage &message);

/*
 * Close the specified mailbox.
 * Returns:
 *    MB_STATUS_OK if the mailbox closed correctly 
 *    MB_STATUS_FAIL_TO_CLOSE otherwise.
 */
MB_RETURN_STATUS MBCloseMailBox(MailBox *mailBox);

/*
 * Checks the mailbox and returns the messages count and size of next message .
 * Returns:
 *    MB_STATUS_OK if the messages count was read ok 
 *    MB_STATUS_NOT_OPEN if the mail box was not opened.
 *    MB_STATUS_ILLEGAL_MODE if the mode of the mail box was not read.
 *    MB_STATUS_ILLEGAL_PARAMS if the mail box was null.
 *    MB_STATUS_CHECK_ERR otherwise.
 */
MB_RETURN_STATUS MBCheckMail(const MailBox *mailBox, LPDWORD messagesCount);

/*
 * Read the next unread message from the mailbox. Do not open the mailbox to read from it!!!
 * Returns:
 *    MB_STATUS_OK if the message was read ok
 *    MB_STATUS_INVALID_MODE if the mailbox open mode is not read.
 *    MB_STATUS_NO_PENDING_MESSAGES if there are no pending messages.
 *    MB_STATUS_ILLEGAL_MSG_FORMAT read message format is illegal.
 *    MB_STATUS_READ_ERR otherwise.
 */
MB_RETURN_STATUS MBReadNextMessage(const MailBox *mailBox, MBMessage *message);

/*
 * Read all the messages in the mailbox and puts them in the messages array.
 * The maxMsgCount is the size of the messages array.
 * Returns:
 *    MB_STATUS_OK if the messages was read ok 
 *    MB_STATUS_INVALID_MODE if the mailbox open mode is not read.
 *    MB_STATUS_NO_PENDING_MESSAGES if there are no pending messages.
 *    MB_STATUS_ILLEGAL_MSG_FORMAT read message format is illegal.
 *    MB_STATUS_READ_ERR otherwise.
 */
MB_RETURN_STATUS MBReadAllMessages(const MailBox *mailBox, 
                                   MBMessage *messages, 
                                   DWORD maxMsgCount, 
                                   LPDWORD readMsgCount);


/*
 * Builds a message from the given parameters.
 * Returns:
 *    MB_STATUS_ILLEGAL_PARAMS if the pMsg is null or the other params exceed their length limit.
 *    MB_STATUS_OK otherwise.
 */
MB_RETURN_STATUS MBMessageBuild(MBMessage* pMsg, 
                                const char* key, 
                                const char* from, 
                                const char* to, 
                                const char* data);

#ifdef __cplusplus
}
#endif	/* __cplusplus */

#endif // _MAIL_BOX_H_
