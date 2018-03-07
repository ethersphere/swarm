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

#include "MailBox.h"

#include <aclapi.h> // to support SetSecurityInfo fix

#define MB_MESSAGE_DELIMITER ','
#define MB_MAIL_SLOT_PREFIX "\\\\.\\mailslot\\"
#define MB_PREFIX_SIZE 14
#define MB_MIN(A,B) (A<B)?(A):(B)

/*
 * Serializes the message for sending to the mailslot.
 */
void MBSerializeMessage(char *msgStr, const MBMessage *msg);

///*
// * Returns the retVal param and free's the buffer if not NULL.
// */
//UINT32
//retfree(UINT32 retVal, VOID* buff)
//{
//	if( buff )
//		free(buff);
//	return( retVal );
//}

/*
 * Parses the string to a message
 *    TRUE - if the parsing succeeded.
 *    FALSE - Otherwise.
 */
BOOL MBParseMessage(MBMessage *msg, const char *msgStr);

void MBMailBoxInit(MailBox *mailBox) {
   mailBox->name[0] = '\0';
   mailBox->handle = NULL;
   mailBox->mode = MB_MODE_WRITE;
}

MB_RETURN_STATUS MBOpenMailBox(MailBox *mailBox, const char *name, MB_MODE mode) {
   HANDLE mailSlotHandle;
   char slotName[MB_NAME_MAX_SIZE + MB_PREFIX_SIZE];
   size_t nameLength;

   // check if the mail box is alread open.
   if (mailBox->handle != NULL) {
      return MB_STATUS_ALREADY_OPEN;
   }

   // check name not null
   if (name == NULL) {
      return MB_STATUS_ILLEGAL_NAME;
   }

   // check name length
   nameLength = strlen(name);
   if (nameLength > MB_NAME_MAX_SIZE) {
      return MB_STATUS_ILLEGAL_NAME;
   }
   strcpy_s(slotName, MB_NAME_MAX_SIZE + MB_PREFIX_SIZE, MB_MAIL_SLOT_PREFIX);
   strcat_s(slotName, MB_NAME_MAX_SIZE + MB_PREFIX_SIZE, name);
   if (mode == MB_MODE_READ) {
      // create the mail slot server.
      mailSlotHandle = CreateMailslotA(slotName, MS_MAX_MESSAGE_SIZE, 0, NULL);

	  if (mailSlotHandle != INVALID_HANDLE_VALUE) // ELAD FIX: allow everyone access to the mailbox
		 SetSecurityInfo(mailSlotHandle,SE_KERNEL_OBJECT, DACL_SECURITY_INFORMATION, NULL,NULL,NULL,NULL );

   }
   else if (mode == MB_MODE_WRITE) {
      // open the mailslot for writing.
      mailSlotHandle = CreateFileA(slotName, 
        GENERIC_WRITE, 
        FILE_SHARE_READ,
        (LPSECURITY_ATTRIBUTES) NULL, 
        OPEN_EXISTING, 
        FILE_ATTRIBUTE_NORMAL, 
        (HANDLE) NULL);
   }
   else {
      return MB_STATUS_INVALID_MODE;
   }
   // check if the file was opened correctly.
   if (mailSlotHandle == INVALID_HANDLE_VALUE) {
      return MB_STATUS_FAIL_TO_OPEN;
   }
   strncpy_s(mailBox->name, MB_NAME_MAX_SIZE + 2, name, nameLength);
   mailBox->mode = mode;
   mailBox->handle = mailSlotHandle;
   return MB_STATUS_OK;
}

MB_RETURN_STATUS MBSendMessage(const MailBox *mailBox, const MBMessage &message) {
   BOOL result;
   DWORD bytesWritten;
   char buffer[MS_MAX_MESSAGE_SIZE];
   // check if the mailbox is open.
   if (mailBox->handle == NULL) {
      return MB_STATUS_MB_NOT_OPEN;
   }
   // check that the mailbox is in write mode.
   if (mailBox->mode != MB_MODE_WRITE) {
      return MB_STATUS_INVALID_MODE;
   }
   MBSerializeMessage(buffer, &message);
   // send the message to the mailslot.
   result = WriteFile(mailBox->handle, buffer, lstrlenA(buffer) + 1, &bytesWritten, NULL);
   if (!result) {
      return MB_STATUS_SEND_ERR;
   }
   return MB_STATUS_OK;
}

MB_RETURN_STATUS MBSendSingleMessage(const char *name, const MBMessage &message) {
   MB_RETURN_STATUS status;
   MailBox mailBox;
   // init the mailbox.
   MBMailBoxInit(&mailBox);

   // open the mailbox.
   status = MBOpenMailBox(&mailBox, name, MB_MODE_WRITE);
   if (status != MB_STATUS_OK) {
      return status;
   }
   // send the message.
   status = MBSendMessage(&mailBox, message);
   if (status != MB_STATUS_OK) {
      MBCloseMailBox(&mailBox); 
      return status;
   }
   // close the mailbox
   status = MBCloseMailBox(&mailBox);
   if (status != MB_STATUS_OK) {
      return status;
   }
   return MB_STATUS_OK;
}

MB_RETURN_STATUS MBCloseMailBox(MailBox *mailBox) {
   // check if already closed.
   if (mailBox->handle == NULL) {
      return MB_STATUS_OK;
   }
   if (CloseHandle(mailBox->handle)) {
      // set the handle to NULL.
      mailBox->handle = NULL;
      return MB_STATUS_OK;
   }
   else {
      return MB_STATUS_FAIL_TO_CLOSE;
   }
}

MB_RETURN_STATUS MBCheckMail(const MailBox *mailBox, LPDWORD messagesCount) {
   if (mailBox == NULL) {
      return MB_STATUS_ILLEGAL_PARAMS;
   }
   if (mailBox->handle == NULL) {
      return MB_STATUS_NOT_OPEN;
   }
   if (mailBox->mode != MB_MODE_READ) {
      return MB_STATUS_INVALID_MODE;
   }
   return (GetMailslotInfo(mailBox->handle, NULL, NULL, messagesCount, NULL))? MB_STATUS_OK : MB_STATUS_CHECK_ERR; 
}

MB_RETURN_STATUS MBReadNextMessage(const MailBox *mailBox, MBMessage *message) {
   DWORD bytesRead, msgCount;
   char buffer[MS_MAX_MESSAGE_SIZE + 2]; // add 2 for null terminition.

   if (mailBox->mode != MB_MODE_READ) {
      return MB_STATUS_INVALID_MODE;
   }
   // check for pending messages.
   MBCheckMail(mailBox, &msgCount);
   if (msgCount <= 0) {
      return MB_STATUS_NO_PENDING_MESSAGES;
   }
   // read the message from the mail slot.
   if (!ReadFile(mailBox->handle, buffer, MS_MAX_MESSAGE_SIZE, &bytesRead, NULL))
      return MB_STATUS_READ_ERR;
   buffer[bytesRead] = '\0';

   if (!MBParseMessage(message, buffer))
      return MB_STATUS_ILLEGAL_MSG_FORMAT;

   return MB_STATUS_OK;
}

MB_RETURN_STATUS MBReadAllMessages(const MailBox *mailBox, 
                                   MBMessage *messages,
                                   DWORD maxMsgCount,
                                   LPDWORD readMsgCount) {
   MB_RETURN_STATUS status;
   DWORD msgCount;
   DWORD msgToRead;
   *readMsgCount = 0;
   if (mailBox->mode != MB_MODE_READ) {
      return MB_STATUS_INVALID_MODE;
   }
   if (maxMsgCount <= 0) {
      return MB_STATUS_ILLEGAL_MSG_COUNT;
   }
   while (1) {
      status = MBCheckMail(mailBox, &msgCount);
      // check the MBCheckMail status
      if (status != MB_STATUS_OK) {
         return MB_STATUS_READ_ERR;
      }
      // check if there are messages left
      if (msgCount == 0) {
         // check if any message was read.
         if (*readMsgCount == 0) {
            return MB_STATUS_NO_PENDING_MESSAGES;
         }
         else {
            return status;
         }
      }
      // check that the messages array is not full
      if (maxMsgCount == *readMsgCount) {
         return MB_STATUS_OK;
      }
      msgToRead = MB_MIN(maxMsgCount - *readMsgCount, msgCount);
      if (msgToRead <=0) {
         return MB_STATUS_OK;
      }
      // read all the messages one by one.
      for (; *readMsgCount < msgToRead; ++(*readMsgCount)) {
         status = MBReadNextMessage(mailBox, messages++);
         if (status != MB_STATUS_OK) {
            return status;
         }
      }
   }
}

/*
 * adds the field and a delimiter to the dest string and advances the string pointer.
 */
void MBAddField(char **dest, const char *field, int maxSize) {
   strncpy_s(*dest,strlen(field) + 1,field, strlen(field) + 1);
   *dest += strlen(field);
   **dest = MB_MESSAGE_DELIMITER;
   (*dest)++;
}

void MBSerializeMessage(char *msgStr, const MBMessage *msg) {
   char *curPtr = msgStr;
   if ((msg == NULL) || (msgStr == NULL)) {
      return;
   }
   MBAddField(&curPtr, msg->key, MB_KEY_MAX_SIZE);
   MBAddField(&curPtr, msg->from, MB_NAME_MAX_SIZE);
   MBAddField(&curPtr, msg->to, MB_NAME_MAX_SIZE);
   MBAddField(&curPtr, msg->data, MB_DATA_MAX_SIZE);
   // add \0.
   curPtr--;
   *curPtr = '\0';
}

/*
 * parses the src string and put the next field in the field param.
 * Returns:
 *    TRUE if the parsing went well.
 *    FALSE if the field's actual size exceeded the max size.
 */
BOOL MBParseField(char *field, const char **src, int maxSize) {
   const char *start = *src;
   const char *end;
   int fieldSize;
   if (*src == NULL) {
      return FALSE;
   }
   // find the delimiter.
   end = strchr(start, MB_MESSAGE_DELIMITER);
   // set the actual field size according to the delimiter location or the end of the string.
   fieldSize = (end == NULL)?(int)strlen(start) : (int)(end-start);
   if (fieldSize > maxSize) {
      return FALSE;
   }
   // copy the field and add null termination.
   strncpy_s(field,maxSize, start, fieldSize);
   field[fieldSize] = '\0';
   // advance the source pointer or null if no other tokens.
   *src = (end == NULL) ? NULL : end + 1;
   return TRUE;
}

BOOL MBParseMessage(MBMessage *msg, const char *msgStr) {
   const char *start = msgStr;
   if (MBParseField(msg->key, &start, MB_KEY_MAX_SIZE) &&
      MBParseField(msg->from, &start, MB_NAME_MAX_SIZE) &&
      MBParseField(msg->to, &start, MB_NAME_MAX_SIZE) &&
      MBParseField(msg->data, &start, MB_DATA_MAX_SIZE)) {
      return TRUE;
   }
   return FALSE;
}

MB_RETURN_STATUS MBMessageBuild(MBMessage* pMsg, 
                                const char* key, 
                                const char* from, 
                                const char* to, 
                                const char* data) {
   if ((pMsg == NULL) || 
      (strlen(key) > MB_KEY_MAX_SIZE) ||
      (strlen(from) > MB_NAME_MAX_SIZE) ||
      (strlen(to) > MB_NAME_MAX_SIZE) ||
      (strlen(data) > MB_DATA_MAX_SIZE)) 
   {
      return MB_STATUS_ILLEGAL_PARAMS;
   }
   strcpy_s(pMsg->key, MB_KEY_MAX_SIZE + 1, key);
   strcpy_s(pMsg->from, MB_NAME_MAX_SIZE+ 1, from);
   strcpy_s(pMsg->to, MB_NAME_MAX_SIZE + 1, to);
   strcpy_s(pMsg->data, MB_DATA_MAX_SIZE + 1, data);
   return MB_STATUS_OK;
}

