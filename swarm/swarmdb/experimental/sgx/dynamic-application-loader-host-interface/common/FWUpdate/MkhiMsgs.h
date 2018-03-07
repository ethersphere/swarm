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

/*
File Name:
   MkhiMsgs.h
Abstract:
   Contains definitions for host message header and commands
Authors:
   Tam Nguyen
   Shiva Janakiraman
*/

#ifndef _MKHI_MSGS_H
#define _MKHI_MSGS_H

#include "MkhiHdrs.h"
#include "pinfo.h"
//#include "mdes.h"


#ifdef _WIN32
#pragma warning (disable: 4214 4200)
#endif // _WIN32

#pragma pack(1)


#define MKHI_MSG_VERSION_MINOR          0x1
#define MKHI_MSG_VERSION_MAJOR          0x1    

#define MKHI_MSG_VERSION                MAKE_PROTOCOL_VERSION(1, 0)

#define MKHI_ACK_BIT                                  BIT7


//Enums for Result field of MHKI Header
#define ME_SUCCESS                     0x00
#define ME_ERROR_ALIAS_CHECK_FAILED    0x01
#define ME_INVALID_MESSAGE             0x02
#define ME_M1_DATA_OLDER_VER           0x03
#define ME_M1_DATA_INVALID_VER         0x04
#define ME_INVALID_M1_DATA             0x05

// Typedef for GroupID 
/** \page groupid MKHI messages Group ID
*Each MKHI client in firmware is identified by its GROUP ID. Following GROUPIDs are supported in firmware: 
*<table border = "3">
<tr>
<td><center><b> Group ID </b></center></td>
<td><center><b>Category Details</b></center></td>
</tr>
<tr>
<td>MKHI_CBM_GROUP_ID=0x00</td>
<td>Core BIOS Messages targeted for PM driver</td> 
</tr>
<tr>
<td>MKHI_PM_GROUP_ID=0x01</td>
<td>PM Config Messages</td> 
</tr> 
<tr>
<td>MKHI_PWD_GROUP_ID=0x02</td>
<td>Password config messages</td> 
</tr> 
<tr>
<td>MKHI_FWCAPS_GROUP_ID=0x03</td>
<td>FW Capabilities config messages</td> 
</tr> 
<tr>
<td>MKHI_APP_GROUP_ID=0x04</td>
<td>Application config data access messages</td> 
</tr> 
<tr>
<td>MKHI_FWUPDATE_GROUP_ID=0x05</td>
<td>Used to Query and resume FW update process. This is used for manufacturing downgrade.</td> 
</tr> 
<tr>
<td>MKHI_FIRMWARE_UPDATE_GROUP_ID=0x06</td>
<td>Used for FW upgrade process. Need to check with FWU team..</td> 
</tr> 
<tr>
<td>MKHI_BIST_GROUP_ID=0x07</td>
<td>Used for perform ME BIST tests.</td> 
</tr> 
<tr>
<td>MKHI_MDES_GROUP_ID=0x08</td>
<td>Used for ME debug MDES messaging </td> 
</tr> 
<tr>
<td>MKHI_ME_DBG_GROUP_ID=0x09</td>
<td>Used internally for ME debug.</td> 
</tr> 
<tr>
<td>MKHI_FPF_GROUP_ID=0x10</td>
<td> Maximum value of MKHI group ID</td> 
</tr> 
<tr>
<td>MKHI_MAX_GROUP_ID=0x11</td>
<td> Maximum value of MKHI group ID</td> 
</tr> 
<tr>
<td>MKHI_GEN_GROUP_ID =0xff</td>
<td>General messages targeted to MKHI fw client</td> 
</tr> 
*</table>  
*/
typedef enum
{
   MKHI_CBM_GROUP_ID = 0,
   MKHI_PM_GROUP_ID, //Reserved (no longer used)
   MKHI_PWD_GROUP_ID,
   MKHI_FWCAPS_GROUP_ID,
   MKHI_APP_GROUP_ID,      // Reserved (no longer used).
   MKHI_FWUPDATE_GROUP_ID, // This is for manufacturing downgrade
   MKHI_FIRMWARE_UPDATE_GROUP_ID,
   MKHI_BIST_GROUP_ID,
   MKHI_MDES_GROUP_ID,
   MKHI_ME_DBG_GROUP_ID,
   MKHI_FPF_GROUP_ID,
   MKHI_MAX_GROUP_ID,
   MKHI_GEN_GROUP_ID = 0xFF
}MKHI_GROUP_ID;

//NOTE!!!!!!!PLEASE READ
//The defines for MKHI COMMANDS have moved down further in this file.
// They are located at line 710.
//IF DOING A PRE 8.0 MERGE TO THIS FILE PLEASE BE AWARE AN DO NOT
// MERGE ANY NEW COMMANDS TO THE TOP OF THE FILE...MOVE THE COMMAND
// SO IT IS CONTIGOUS WITH THE MKHI COMMANDS DEFINES AT 710.
//NOTE!!!!!!!!PLEASE READ

#define MKHI_IS_GROUP_ID_NOT_USED(gid) ((((gid) == MKHI_APP_GROUP_ID) || ((gid) == MKHI_PM_GROUP_ID)) ? TRUE : FALSE )  
//Number of clients expected to register is MKHI_MAX_GROUP_ID - 2
// The 2 are MKHI_APP_GROUP_ID (no handler) and MKHI_PM_GROUP_ID (deprecated)
#define MKHI_NUM_CLIENTS_EXPECTED 9 
C_ASSERT(MKHI_NUM_CLIENTS_EXPECTED == (MKHI_MAX_GROUP_ID - 2));

#define MKHI_FW_UPDATE_GROUP_ID        MKHI_FIRMWARE_UPDATE_GROUP_ID // Maintain sownward compatability

#pragma pack(1)

//messages definition that HCI handles
/** \addtogroup   mkhiversion 
 * @{
 *    This command is used to get MKHI version number. \n 
 *    <b> Data Structure: </b> \n 
 *    \li <b>_GEN_GET_MKHI_VERSION </b>
 *    \li <b>_GEN_GET_MKHI_VERSION_ACK</b>
 *    \li <b>_MKHI_VERSION</b>
 *
 * @} 
 */
typedef union _MKHI_VERSION
{
   UINT32    Data;
   struct
   {
      UINT32 Minor :16;
      UINT32 Major :16;
   }Fields;
}MKHI_VERSION; 

//MKHI version messages
typedef struct _GEN_GET_MKHI_VERSION
{
   MKHI_MESSAGE_HEADER  Header;
}GEN_GET_MKHI_VERSION;

typedef struct _GET_MKHI_VERSION_ACK_DATA
{
   MKHI_VERSION   MKHIVersion;
}GET_MKHI_VERSION_ACK_DATA;

typedef struct _GEN_GET_MKHI_VERSION_ACK
{
   MKHI_MESSAGE_HEADER        Header;
   GET_MKHI_VERSION_ACK_DATA  Data;
}GEN_GET_MKHI_VERSION_ACK; 

/** \addtogroup   fwversion
 * @{
 *    This command is used to get FW version number. \n 
 *    <b> Data Structure: </b> \n 
 *    \li <b>_GEN_GET_FW_VERSION</b>
 *    \li <b>_GEN_GET_FW_VERSION_ACK</b>
 *    \li <b>_FW_VERSION</b>
 *
 * @} 
 */  
typedef struct _FW_VERSION
{
   UINT32 CodeMinor   :16;
   UINT32 CodeMajor   :16;
   UINT32 CodeBuildNo :16;
   UINT32 CodeHotFix  :16;
   UINT32 NFTPMinor   :16;
   UINT32 NFTPMajor   :16;
   UINT32 NFTPBuildNo :16;
   UINT32 NFTPHotFix  :16;
   UINT32 FITCMinor   :16;
   UINT32 FITCMajor   :16;
   UINT32 FITCBuildNo :16;
   UINT32 FITCHotFix  :16;
}FW_VERSION;

//FW version messages
typedef struct _GEN_GET_FW_VERSION
{
   MKHI_MESSAGE_HEADER  Header;
}GEN_GET_FW_VERSION;

typedef struct _GET_FW_VERSION_ACK_DATA
{
   FW_VERSION  FWVersion;

}GET_FW_VERSION_ACK_DATA;

typedef struct _GEN_GET_FW_VERSION_ACK
{
   MKHI_MESSAGE_HEADER      Header;
   GET_FW_VERSION_ACK_DATA  Data;
}GEN_GET_FW_VERSION_ACK;

//Unconfig without password messages
typedef struct _GEN_UNCFG_WO_PWD
{
   MKHI_MESSAGE_HEADER  Header;
}GEN_UNCFG_WO_PWD;

typedef struct _GEN_UNCFG_WO_PWD_ACK
{
   MKHI_MESSAGE_HEADER  Header;
}GEN_UNCFG_WO_PWD_ACK;

/** \addtogroup eopblock
*<b>END OF POST</b>\n
During BIOS POST, ME allows certain operations to take place that, during normal functionality, would require to meet highersecurity requirements.This works under basic premise that BIOS is considered more secure than software running under host-OS.Examples of such interfaces are AMTHI (Intel Active Management Technology Host Interface), FSC (Fan Speed Control) interface etc. These interfaces are closed once BIOS informs ME that it is ready to load the host OS or EFI shell. Although EFI shell is technically still part of BIOS, the EFI shell is considered less secure then BIOS. EFI shell allows users to run software that is not validated and approved by the BIOS manufacturer. Therefore the End of Post message should be sent before EFI shell is entered. Note that this message is ONLY sent if ME is working in NORMAL mode of operation. This means that if ME is DISABLED,in ERROR state, in RECOVERY mode of operation then this message is NOT sent.   

<b>END OF POST ACK</b>\n
This message is sent by the ME to the host in response to the END OF POST message.  BIOS can proceed only after receiving this response.BIOS does not wait for a response when making S3-exit. In S4/5-exit or G3-exit, BIOS waits to receive a response within 5 seconds.If it does not receive a response, it should halt with a warning message to user.\n

BIOS must wait for a response for the END_OF_POST request message because there is a boundary case where host s/w starts loading before ME has processed this message.In this case, a rogue host s/w (e.g. boot virus) can potentially overwrite the END OF POST message before it has been serviced by ME by continuously writing content in HECI circular buffer. HECI H/W does not provide protection against overrun and if the write pointer reaches the location in HECI buffer where END OF POST message is written then it will be overwritten by next write. 
In case of GRST requested <b>_CBM_EOP_ACK_DATA</b> is populated with  EOP_DATA_PERFORM_GLOBAL_RESET = 1.

\code
             Response Message Required: Yes, CBM_EOP_ACK_DATA
\endcode
* <b> Data Structure: </b>
*  \li <b>  _GEN_END_OF_POST</b>
*  \li <b>  _GEN_END_OF_POST_ACK</b> 
*  \li <b>  _CBM_EOP_ACK_DATA </b> 
*
* This command is sent under MKHI group ID  MKHI_GEN_GROUP_ID. 
*/

//End of Post message data defns
#define EOP_DATA_STATUS_SUCCESS          0x0
#define EOP_DATA_PERFORM_GLOBAL_RESET    0x1    

//End of Post message data
typedef struct _CBM_EOP_ACK_DATA
{
   UINT32 RequestedActions;
}CBM_EOP_ACK_DATA;

//End of POST message
typedef struct _GEN_END_OF_POST
{
   MKHI_MESSAGE_HEADER  Header;
}GEN_END_OF_POST;

//End of POST ack message
typedef struct _GEN_END_OF_POST_ACK
{
   MKHI_MESSAGE_HEADER  Header;
   CBM_EOP_ACK_DATA Data;
}GEN_END_OF_POST_ACK;

//Get ME unconfigure state
typedef struct _GEN_GET_ME_UNCFG_STATE
{
   MKHI_MESSAGE_HEADER  Header;
}GEN_GET_ME_UNCFG_STATE;

//Get ME unconfigure state ack
typedef struct _GEN_GET_ME_UNCFG_STATE_ACK
{
   MKHI_MESSAGE_HEADER  Header;
}GEN_GET_ME_UNCFG_STATE_ACK;

typedef struct _GEN_UPDATE_CPU_PINFO_DATA
{
   UINT8	   CommandCode;				
   UINT16	CommandDataSize;			
   PINFO 	CommandData; 
} GEN_UPDATE_CPU_PINFO_DATA;

// Update CPU ID message
typedef struct _GEN_UPDATE_CPU_PINFO
{
   MKHI_MESSAGE_HEADER         Header;
   GEN_UPDATE_CPU_PINFO_DATA	 Data;
} GEN_UPDATE_CPU_PINFO;
// Update CPU ID ack
typedef struct _GEN_UPDATE_CPUID_ACK
{
    MKHI_MESSAGE_HEADER    Header;
} GEN_UPDATE_CPUID_ACK;

// Update CPU ID ack
typedef struct _GEN_UPDATE_CPU_PINFO_ACK
{
    MKHI_MESSAGE_HEADER 	Header;
} GEN_UPDATE_CPU_PINFO_ACK;



//	Get CPU Type Change  
/**	\addtogroup cpubrand 
* @{
*The following messages are used by MEBx to query ME FW whether MEBx is supposed to launch End User Interaction regarding CPU Replacement. \n
*
* <b> Data structure: </b> \n
*  \li <b>_GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE </b>
*  \li <b>_GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK_DATA</b>
*  \li <b>_GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK</b>
*  \li <b>_GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_DATA</b>
*  \li <b>_GEN_SEND_CPU_BRAND_CLASS_FUSE</b>
*  \li <b>_GEN_SEND_CPU_BRAND_CLASS_FUSE_ACK</b> 
*
* @}
*/
typedef struct _GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE
{
   MKHI_MESSAGE_HEADER      Header;
} GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE;

typedef enum _USER_FEEDBACK_REQUEST
{
   USER_FEEDBACK_NOT_REQUESTED = 0,
   USER_FEEDBACK_REQUESTED
} USER_FEEDBACK_REQUEST;
/**	\addtogroup cpubrand 
   * <b>_GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK</b> command has the following attribute.
   * \code 
   *   UserFeedback 
   *                    USER_FEEDBACK_NOT_REQUESTED (=0)
   *                    USER_FEEDBACK_REQUESTED (=1) 
   *   FeaturesDisabled  
   *                    UINT32 bitmap of features being disabled with the CPU CHANGE.
   *                    If UserFeedback equals 0 ( USER_FEEDBACK_NOT_REQUESTED) then ignore the field. 
   *   FeaturesEnabled  
   *                    UINT32 bitmap of features being enabled with the CPU CHANGE.
   *                    If UserFeedback equals 0 ( USER_FEEDBACK_NOT_REQUESTED) then ignore the field.
   *          
   *   GlobalResetRequired 
   *                    Optional Parameter 
   *                    If UserFeedback equals 1 ( USER_FEEDBACK_REQUESTED) then ignore the field. 
   *                    Otherwise trigger Global Reset if this field is set to 1.
   *           
   * \endcode
   */
//	Get CPU Type Change  Ack
typedef struct _GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK_DATA

{
   UINT8          UserFeedback;
   UINT32         FeaturesDisabled;
   UINT32         FeaturesEnabled;
   UINT8          GlobalResetRequired;
} GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK_DATA; 

typedef struct _GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK
{
   MKHI_MESSAGE_HEADER      		                   Header;
   GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK_DATA	 Data;
} GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_ACK;

//	Set CPU Type Change User Response
typedef enum _CPU_TYPE_CHANGE_USER_RESPONSE
{
   CPU_TYPE_CHANGE_REJECTED = 0,
   CPU_TYPE_CHANGE_ACCEPTED
} CPU_TYPE_CHANGE_USER_RESPONSE;
/** \addtogroup cpubrand 
    * <b>_GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE</b>command has the following attribute. 
    * \code
    *  UserResponse =  CPU_TYPE_CHANGE_REJECTED (=0) 
    *                  CPU_TYPE_CHANGE_ACCEPTED (=1)
    * \endcode
    *                  
*/
typedef struct _GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_DATA
{
   UINT8                    UserResponse;
} GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_DATA;

typedef struct _GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE
{
   MKHI_MESSAGE_HEADER      Header;
   GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_DATA Data;
} GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE;

// Set CPU Type Change User Response Ack
typedef struct _GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_ACK
{
   MKHI_MESSAGE_HEADER      		          Header;
} GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_ACK;

////////////////////////////////////////////////////////////////////////////////
// New CPU type definitions
typedef enum _CPU_BRAND_TYPE
{
   CPU_TYPE_UNIDENTIFIED = 0,
   CPU_TYPE_CORE2_NONVPRO,
   CPU_TYPE_VPRO,
   CPU_TYPE_PENTIUM,
   CPU_TYPE_CELERON,
   CPU_TYPE_XEON,
   CPU_TYPE_XEONVPRO,
   CPU_TYPE_DONT_CARE = 0xFF
} CPU_BRAND_TYPE;



// Send CPU Brand Class Fuse msg

typedef struct _GEN_SEND_CPU_BRAND_CLASS_FUSE
{
   MKHI_MESSAGE_HEADER                  Header;
   UINT8 CpuBrandClass;           /* CPU Brand Class value read from CPU fuses */
} GEN_SEND_CPU_BRAND_CLASS_FUSE;

// Update CPU ID ack
typedef struct _GEN_SEND_CPU_BRAND_CLASS_FUSE_ACK
{
    MKHI_MESSAGE_HEADER Header;
} GEN_SEND_CPU_BRAND_CLASS_FUSE_ACK;


// Vpro allowed Cmd  :Lenovo RCR 
typedef enum  _VPRO_ALLOWED_STATE
{
   VPRO_NOT_ALLOWED =0,
   VPRO_ALLOWED
} VPRO_ALLOWED_STATE;

// this will come from BIOS to set the vpro allowed nvar
typedef struct _GEN_SET_VPRO_ALLOWED
{
   MKHI_MESSAGE_HEADER Header;
   UINT8      VproState;
} GEN_SET_VPRO_ALLOWED;

//Set Vpro Allowed State response back from FW
typedef struct _GEN_SET_VPRO_ALLOWED_ACK
{
   MKHI_MESSAGE_HEADER Header;
} GEN_SET_VPRO_ALLOWED_ACK;

// Query Vpro Allowed nvar State
typedef struct _GEN_GET_VPRO_ALLOWED{
   MKHI_MESSAGE_HEADER Header;
} GEN_GET_VPRO_ALLOWED;

typedef struct _GEN_GET_VPRO_ALLOWED_ACK
{
   MKHI_MESSAGE_HEADER Header;
   UINT8   VproState;   
} GEN_GET_VPRO_ALLOWED_ACK;


typedef struct _GEN_GET_ROM_BIST_DATA
{
   MKHI_MESSAGE_HEADER Header;
}GEN_GET_ROM_BIST_DATA;


// DBG UMCHID for LPT with secuirty, PAVP and GID all zero
// Revisit this for other platform
#define DBG_UMCHID  {0x92,0x5c, 0x18, 0xf4, 0x85, 0x61, 0x8e, 0xc1, 0xdf, 0x65, 0x2a, 0x2b, 0xa4, 0x64, 0xfd, 0x0e}
typedef struct _GEN_GET_ROM_BIST_DATA_ACK_DATA
{
   UINT16   DeviceId;
   UINT16   FuseTestFlags;
   UINT8    Umchid_hash[12]; 
   UINT32   Rand; 
#if DBG
   UINT32   Umchid[4]; 
#endif
}GEN_GET_ROM_BIST_DATA_ACK_DATA;

typedef struct _GEN_GET_ROM_BIST_DATA_ACK
{
   MKHI_MESSAGE_HEADER              Header;
   GEN_GET_ROM_BIST_DATA_ACK_DATA   Data;
}GEN_GET_ROM_BIST_DATA_ACK;
/** \addtogroup mfgmrstblock 
* @{
This message is sent by tools to FW in the manufacturing environment.This reset request generates a ME only reset and causesFW to halt in bring-up.This command is sent under group ID MKHI_GEN_GROUP_ID.
\code
    Response Message Required:	yes
    NOTE: If this message is sent after END_OF_POST message, it is denied and an error 
          response is sent.
\endcode
 
<b>Data structure: </b>\n
\li <b>_MKHI_MESSAGE_HEADER</b>
\li <b>_GEN_SET_MFG_MRST_AND_HALT_ACK</b>

* @}
*/
typedef struct _GEN_SET_MFG_MRST_AND_HALT_ACK
{
   MKHI_MESSAGE_HEADER  Header;   
}GEN_SET_MFG_MRST_AND_HALT_ACK;


typedef struct{
   UINT32 FileName;
   struct {
      UINT32 IsBlob : 1;
      UINT32 GetDefault : 1;
      UINT32 NvarHash:1;
      UINT32 Reserved : 29;
   }Fields;

   UINT32 FileReturnSize;
   UINT32 FileReadOffset;
} GET_FILE_REQ_DATA;

typedef struct{
   UINT32 FileSize;
   UINT8 File[1];
} GET_FILE_ACK_DATA;

typedef struct _GEN_GET_FILE_REQ
{
   MKHI_MESSAGE_HEADER Header;
   GET_FILE_REQ_DATA      Data;
}GEN_GET_FILE_REQ;


typedef struct _GEN_GET_FILE_ACK
{
   MKHI_MESSAGE_HEADER   Header;
   GET_FILE_ACK_DATA     Data;
}GEN_GET_FILE_ACK;

typedef struct _GEN_UPDATE_DEFAULTS_ACK
{
   MKHI_MESSAGE_HEADER              Header;
}GEN_UPDATE_DEFAULTS_ACK;


typedef struct _GEN_SET_FEATURE_STATE_DATA
{
   UINT32 EnableFeature;   //  MEFWCAPS_SKU
   UINT32 DisableFeature;  //MEFWCAPS_SKU
} GEN_SET_FEATURE_STATE_DATA;

typedef struct _GEN_SET_FEATURE_STATE
{
   MKHI_MESSAGE_HEADER      Header;
   GEN_SET_FEATURE_STATE_DATA FeatureState;
} GEN_SET_FEATURE_STATE;


typedef enum _SET_FEATURE_STATE_RESPONSE
{
   SET_FEATURE_STATE_ACCEPTED = 0,
   SET_FEATURE_STATE_REJECTED
} SET_FEATURE_STATE_RESPONSE; 

typedef struct _GEN_SET_FEATURE_STATE_ACK_DATA
{
   SET_FEATURE_STATE_RESPONSE  Response;
} GEN_SET_FEATURE_STATE_ACK_DATA;

typedef struct _GEN_SET_FEATURE_STATE_ACK
{
   MKHI_MESSAGE_HEADER      Header;
   GEN_SET_FEATURE_STATE_ACK_DATA Data;
} GEN_SET_FEATURE_STATE_ACK;

typedef struct _GEN_GET_IMAGE_TYPE
{
	MKHI_MESSAGE_HEADER     Header;
}GEN_GET_IMAGE_TYPE;

typedef struct _GEN_GET_IMAGE_TYPE_ACK_DATA
{
	UINT32	IsProduction;
}GEN_GET_IMAGE_TYPE_ACK_DATA;

typedef struct _GEN_GET_IMAGE_TYPE_ACK
{
	MKHI_MESSAGE_HEADER	            Header;
	GEN_GET_IMAGE_TYPE_ACK_DATA     Data;
}GEN_GET_IMAGE_TYPE_ACK;

typedef struct _GEN_GET_PCH_TYPE
{
	MKHI_MESSAGE_HEADER             Header;
}GEN_GET_PCH_TYPE;


typedef struct _GEN_GET_PCH_TYPE_ACK_DATA
{
	UINT32  IsProduction;
	UINT32  IsSuperSku;
}GEN_GET_PCH_TYPE_ACK_DATA;

typedef struct _GEN_GET_PCH_TYPE_ACK
{
	MKHI_MESSAGE_HEADER            Header;
	GEN_GET_PCH_TYPE_ACK_DATA      Data;
}GEN_GET_PCH_TYPE_ACK;
//set system integrator Id message
typedef struct _SET_SYSTEM_INTEGRATOR_ID_DATA
{
   UINT32   SysIntId;
   UINT8    Index;
}SET_SYSTEM_INTEGRATOR_ID_DATA;

typedef struct _SET_SYSTEM_INTEGRATOR_ID
{
   MKHI_MESSAGE_HEADER           Header;
   SET_SYSTEM_INTEGRATOR_ID_DATA Data;
}SET_SYSTEM_INTEGRATOR_ID;

typedef struct _SET_SYSTEM_INTEGRATOR_ID_ACK
{
   MKHI_MESSAGE_HEADER  Header;
}SET_SYSTEM_INTEGRATOR_ID_ACK;


//get system integrator Id message
typedef struct _GET_SYSTEM_INTEGRATOR_ID_DATA
{
   UINT8  Index;
}GET_SYSTEM_INTEGRATOR_ID_DATA;
typedef struct _GET_SYSTEM_INTEGRATOR_ID
{
   MKHI_MESSAGE_HEADER            Header;
   GET_SYSTEM_INTEGRATOR_ID_DATA  Data;
}GET_SYSTEM_INTEGRATOR_ID;

typedef struct _GET_SYSTEM_INTEGRATOR_ID_ACK_DATA
{
   UINT32   SysIntId;
}GET_SYSTEM_INTEGRATOR_ID_ACK_DATA;

typedef struct _GET_SYSTEM_INTEGRATOR_ID_ACK
{
   MKHI_MESSAGE_HEADER                 Header;
   GET_SYSTEM_INTEGRATOR_ID_ACK_DATA   Data;
}GET_SYSTEM_INTEGRATOR_ID_ACK;

typedef struct _GET_INVOCATION_CODE
{
    MKHI_MESSAGE_HEADER    Header;
} GET_INVOCATION_CODE;

typedef struct _GET_INVOCATION_CODE_ACK
{
    MKHI_MESSAGE_HEADER    Header;
    UINT32                 InvocationCode;
} GET_INVOCATION_CODE_ACK;

typedef struct _SET_INVOCATION_CODE
{
    MKHI_MESSAGE_HEADER    Header;
    UINT32                 InvocationCode;
} SET_INVOCATION_CODE;

typedef struct _SET_INVOCATION_CODE_ACK
{
    MKHI_MESSAGE_HEADER    Header;
} SET_INVOCATION_CODE_ACK;

typedef struct _CLR_INVOCATION_CODE
{
    MKHI_MESSAGE_HEADER    Header;
    UINT32                 InvocationCode;
} CLR_INVOCATION_CODE;

typedef struct _CLR_INVOCATION_CODE_ACK
{
    MKHI_MESSAGE_HEADER    Header;
} CLR_INVOCATION_CODE_ACK;

typedef struct _PWR_GATE_ME_REQ
{
   MKHI_MESSAGE_HEADER    Header;
} PWR_GATE_ME_REQ;

typedef struct _GET_SPG_STATUS
{
   MKHI_MESSAGE_HEADER    Header;
} GET_SPG_STATUS;

typedef union _GET_SPG_STATUS_ACK_DATA
{
   struct
   {
      UINT32 BlockedByFwSku         :1;
      UINT32 BlockedByHwSku         :1;
      UINT32 BlockedByOemOverride   :1;
      UINT32 BlockedByUserOverride  :1;
      UINT32 BlockedByBiosOverride  :1;
      UINT32 BlockedByFwNotReady    :1;
   } b;
   UINT32 ul;
} GET_SPG_STATUS_ACK_DATA;

typedef struct _GET_SPG_STATUS_ACK
{
   MKHI_MESSAGE_HEADER     Header;
   GET_SPG_STATUS_ACK_DATA Status;
} GET_SPG_STATUS_ACK;
C_ASSERT(sizeof(GET_SPG_STATUS_ACK) == 8);

#pragma pack() 



/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///HCI GENERIC COMMAND
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//command handle by HCI 
/** \addtogroup   mkhiversion  
 * \code 
 * This command is sent under group id MKHI_GEN_GROUP_ID.
 * Command request  GEN_GET_MKHI_VERSION_CMD     =  0x01 
 * Command response GEN_GET_MKHI_VERSION_CMD_ACK =  0x81 
 * \endcode
 */
#define GEN_GET_MKHI_VERSION_CMD        0x01
#define GEN_GET_MKHI_VERSION_CMD_ACK                  (MKHI_ACK_BIT | GEN_GET_MKHI_VERSION_CMD)
/** \addtogroup   fwversion  
 * \code
 * This command is sent under group id MKHI_GEN_GROUP_ID.
 * Command request  GEN_GET_FW_VERSION_CMD     = 0x02  
 * Command reponse  GEN_GET_FW_VERSION_CMD_ACK = 0x82  
 * \endcode
 */
#define GEN_GET_FW_VERSION_CMD          0x02
#define GEN_GET_FW_VERSION_CMD_ACK                    (MKHI_ACK_BIT | GEN_FW_VERSION_CMD)

#define GEN_RESERVED1_CMD       0x03     //was #define GEN_SET_MEBX_BIOS_VER_CMD  0x03
#define GENT_RESERVED1_CMD_ACK                        (MKHI_ACK_BIT | GEN_RESERVED1_CMD)

#define GEN_UPDATE_DEFAULTS_CMD                0x04
#define GEN_UPDATE_DEFAULTS_CMD_ACK                   (MKHI_ACK_BIT | GEN_UPDATE_DEFAULTS_CMD)

#define GEN_UPDATE_CPUID_CMD		    0x05
#define GEN_UPDATE_CPUID_CMD_ACK                      (MKHI_ACK_BIT | GEN_UPDATE_CPUID_CMD)

#define GEN_UPDATE_PINFO_CMD                 0x06
#define GEN_UPDATE_PINFO_CMD_ACK                      (MKHI_ACK_BIT | GEN_UPDATE_PINFO_CMD)
/** \addtogroup cpubrand 
* \code  
* Commands request and reponses for CPU Brand class messsages are sent under group id MKHI_GEN_GROUP_ID
* Commands request and reponses for CPU Brand class messsages are:
*
*     GEN_SEND_CPU_BRAND_CLASS_FUSE_CMD     = 0x07  
*     GEN_SEND_CPU_BRAND_CLASS_FUSE_CMD_ACK = 0x87 
*
*     GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_CMD     =  0x08
*     GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_CMD_ACK =  0x88
*
*     GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_CMD     =  0x09  
*     GEN_SET_CPU_TYPE_CHANGE_USER_REPSONSE_CMD_ACK =  0x89 
*     
* \endcode
*/
#define GEN_SEND_CPU_BRAND_CLASS_FUSE_CMD    0x07
#define GEN_SEND_CPU_BRAND_CLASS_FUSE_CMD_ACK         (MKHI_ACK_BIT | GEN_SEND_CPU_BRAND_CLASS_FUSE_CMD)


#define GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_CMD 0x08
#define GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_CMD_ACK  (MKHI_ACK_BIT | GEN_GET_CPU_TYPE_CHANGE_USER_MESSAGE_CMD)

#define GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_CMD  0x09
#define GEN_SET_CPU_TYPE_CHANGE_USER_REPSONSE_CMD_ACK (MKHI_ACK_BIT | GEN_SET_CPU_TYPE_CHANGE_USER_RESPONSE_CMD)

#define GEN_COMMAND_UNUSED_1_CMD                      0x0A   //Not used.
#define GEN_COMMAND_UNUSED_1_ACK                      (MKHI_ACK_BIT | GEN_COMMAND_UNUSED_1_CMD)

#define GEN_GET_FILE_CMD                0x0B
#define GEN_GET_FILE_CMD_ACK                          (MKHI_ACK_BIT | GEN_GEET_FILE_CMD)
/** \addtogroup eopblock   
* \code
* Command request:  GEN_END_OF_POST_CMD     =  0x0C  
* Command response: GEN_END_OF_POST_CMD_ACK =  0x8C 
* \endcode
*/
#define GEN_END_OF_POST_CMD             0x0C
#define GEN_END_OF_POST_CMD_ACK                       (MKHI_ACK_BIT | GEN_END_OF_POST_CMD)

#define GEN_UNCFG_WO_PWD_CMD            0x0D
#define GEN_UNCFG_WO_PWD_CMD_ACK                      (MKHI_ACK_BIT | GEN_UNCFG_WO_PWD_CMD)

#define GEN_GET_ME_UNCFG_STATE_CMD      0x0E
#define GEN_GET_ME_UNCFG_STATE_CMD_ACK                (MKHI_ACK_BIT | GEN_GET_ME_UNCFG_STATE_CMD)

#define GEN_GET_ROM_BIST_DATA_CMD       0x0F
#define GEN_GET_ROM_BIST_DATA_CMD_ACK                 (MKHI_ACK_BIT | GENT_GET_ROM_BIST_DATA_CMD)
/** \addtogroup mfgmrstblock 
* \code
* Commmand request: GEN_SET_MFG_MRST_AND_HALT_CMD      0x10 
* Commmand response:GEN_SET_MFG_MRST_AND_HALT_CMD_ACK  0x90  
* \endcode
*/ 
#define GEN_SET_MFG_MRST_AND_HALT_CMD   0x10
#define GEN_SET_MFG_MRST_AND_HALT_CMD_ACK             (MKHI_ACK_BIT | GEN_SET_MFG_MRST_AND_HANDL_CMD)
/** \addtogroup memaddressblock 
  This commmand is sent up MKHI group  MKHI_GEN_GROUP_ID. \n 
  Command requests and responses are: 
  \code
      GEN_SET_MEMORY_ADDRESS_CMD             0x11 
      GEN_SET_MEMORY_ADDRESS_CMD_ACK         0x91 
      GEN_GET_MEMORY_ADDRESS_CMD             0x12 
      GEN_GET_MEMORY_ADDRESS_CMD_ACK         0x92 
 \endcode
*/
#define GEN_SET_MEMORY_ADDRESS_CMD             0x11
#define GEN_SET_MEMORY_ADDRESS_CMD_ACK                (MKHI_ACK_BIT | GEN_COMMAND_MEMORY_ADDRESS_CMD)

#define GEN_GET_MEMORY_ADDRESS_CMD             0x12
#define GEN_GET_MEMORY_ADDRESS_CMD_ACK                (MKHI_ACK_BIT | GEN_GET_MEMORY_ADDRESS_CMD)

#define GEN_SET_SYSTEM_INTEGRATOR_ID_CMD              0x13
#define GEN_SET_SYSTEM_INTEGRATOR_ID_CMD_ACK         (MKHI_ACK_BIT | GEN_SET_SYSTEM_INTEGRATOR_ID_CMD)

#define GEN_SET_FEATURE_STATE_CMD                     0x14  // This is not used currently.
                                                       // RCR 1023363 : Manageability State Control by BIOS.
                                                       // Need BIOS version with the above RCR changes for this to be invoked
#define GEN_SET_FEATURE_STATE_CMD_ACK                 (MKHI_ACK_BIT | GEN_SET_FEATURE_STATE_CMD)

#define GEN_GET_SYSTEM_INTEGRATOR_ID_CMD              0x15
#define GEN_GET_SYSTEM_INTEGRATOR_ID_CMD_ACK          (MKHI_ACK_BIT | GEN_GET_SYSTEM_INTEGRATOR_ID_CMD)

#define GEN_GET_VPRO_ALLOWED_CMD                      0x16
#define GEN_GET_VPRO_ALLOWED_CMD_ACK                  (MKHI_ACK_BIT | GEN_GET_VPRO_ALLOWED_CMD)

#define GEN_SET_VPRO_ALLOWED_CMD                      0x17
#define GEN_SET_VPRO_ALLOWED_CMD_ACK                  (MKHI_ACK_BIT | GEN_SET_VPRO_ALLOWED_CMD)


#define GEN_GET_IMAGE_TYPE_CMD                          0x18
#define GEN_GET_PCH_TYPE_CMD_ACK                        (MKHI_ACK_BIT | GEN_GET_PCH_TYPE_CMD)
#define GEN_GET_PCH_TYPE_CMD                            0x19
#define GEN_GET_IMAGE_TYPE_CMD_ACK                      (MKHI_ACK_BIT | GEN_GET_IMAGE_TYPE_CMD)




///////////// MDES MKHI Command message structure for group ID MKHI_MDES_GROUP_ID ///////////////////////////
// Commands handle directly by HCI as part of MDES group ID (MKHI_MDES_GROUP_ID).
#define  MDES_RAM_LOG_IDENTIFIER                   0
#define  MDES_FLASH_LOG_IDENTIFIER                 1
#define  MDES_MAX_GET_LOG_DATA_SIZE                (MAX_MDES_BUFFER_SIZE) // Need to double check on this?

#define  MDES_GET_VERSION_MKHI_CMD               0x01
#define  MDES_GET_VERSION_MKHI_CMD_ACK           (MKHI_ACK_BIT | MDES_GET_VERSION_MKHI_CMD)

#define  MDES_GET_CONFIG_MKHI_CMD                0x02
#define  MDES_GET_CONFIG_MKHI_CMD_ACK            (MKHI_ACK_BIT | MDES_GET_CONFIG_MKHI_CMD)

#define  MDES_SET_CONFIG_MKHI_CMD                0x03
#define  MDES_SET_CONFIG_MKHI_CMD_ACK            (MKHI_ACK_BIT | MDES_SET_CONFIG_MKHI_CMD)

#define  MDES_PAUSE_LOGGING_MKHI_CMD             0x04
#define  MDES_PAUSE_LOGGING_MKHI_CMD_ACK         (MKHI_ACK_BIT | MDES_PAUSE_LOGGING_MKHI_CMD)

#define  MDES_UNPAUSE_LOGGING_MKHI_CMD           0x05
#define  MDES_UNPAUSE_LOGGING_MKHI_CMD_ACK       (MKHI_ACK_BIT | MDES_UNPAUSE_LOGGING_MKHI_CMD)

#define  MDES_CLEAR_LOG_MKHI_CMD                 0x06
#define  MDES_CLEAR_LOG_MKHI_CMD_ACK             (MKHI_ACK_BIT | MDES_CLEAR_LOG_MKHI_CMD)


#define  MDES_GET_LOG_SIZE_MKHI_CMD              0x07
#define  MDES_GET_LOG_SIZE_MKHI_CMD_ACK          (MKHI_ACK_BIT | MDES_GET_LOG_SIZE_MKHI_CMD)

#define  MDES_GET_LOG_DATA_MKHI_CMD              0x08
#define  MDES_GET_LOG_DATA_MKHI_CMD_ACK          (MKHI_ACK_BIT | MDES_GET_LOG_DATA_MKHI_CMD)

#define  DEBUG_CAPABILITY_ENABLE_MKHI_CMD        0x09
#define  DEBUG_CAPABILITY_ENABLE_MKHI_CMD_ACK    (MKHI_ACK_BIT | DEBUG_CAPABILITY_ENABLE_MKHI_CMD)

#define  DEBUG_CAPABILITY_DISABLE_MKHI_CMD       0x0A
#define  DEBUG_CAPABILITY_DISABLE_MKHI_CMD_ACK   (MKHI_ACK_BIT | DEBUG_CAPABILITY_DISABLE_MKHI_CMD)  

/** \addtogroup  MdesMessages 
*BIOS Messages over MDES will use the following command request and response using MKHI group id <b>MKHI_MDES_GROUP_ID</b>.The command <b>_CBM_BIOS_MDES_MSG_REQ </b> will have no respose or acknowledgement from ME. The command <b>_MKHI_CBM_BIOS_MDES_MSG_GET_CONFIG_REQ</b> will acknowledged with MDES configuration information if successful or will be NACK-ed on error.
\code 
     Command request  used for _CBM_BIOS_MDES_MSG_REQ  is:
                                            MDES_BIOS_MSG_LOG_REQ_CMD=0x0B 

     Command request  used for _MKHI_CBM_BIOS_MDES_MSG_GET_CONFIG_REQ is:
                                            MDES_BIOS_MSG_GET_CONFIG_CMD= 0x0C
     Command response used for _MKHI_CBM_BIOS_MDES_MSG_GET_CONFIG_ACK is:
                                            MDES_BIOS_MSG_GET_CONFIG_ACK = 0x8C 

\endcode
 */

#define MDES_BIOS_MSG_LOG_REQ_CMD                0x0B
#define MDES_BIOS_MSG_LOG_REQ_CMD_ACK            (MKHI_ACK_BIT | MDES_BIOS_MSG_LOG_REQ_CMD)  

#define MDES_BIOS_MSG_GET_CONFIG_CMD             0x0C
#define MDES_BIOS_MSG_GET_CONFIG_ACK             (MKHI_ACK_BIT | MDES_BIOS_MSG_LOG_REQ_CMD) 

//Unconfig without password is in progress
#define ME_UNCONFIG_IN_PROGRESS        0x01
//normal case, there is unconfigure w/o password to be processed
#define ME_UNCONFIG_NOT_IN_PROGRESS    0x02
//Me returns this status to bios when it finished processing unconfig 
//w/o password. When bios see this status, it will perform a global reset.
#define ME_UNCONFIG_FINISHED           0x03
//Me encountered error while processing revert back to default.
//any specific error status must be defined starting at 0x81
#define ME_UNCONFIG_ERROR              0x80
//Used to track HCI's first boot.
#define ME_UNCONFIG_FIRST_BOOT         0x81
//Used to indicate to HCI that it is
// no longer on first boot.
#define ME_UNCONFIG_NOT_FIRST_BOOT         0x82

#endif //_MKHI_MSGS_H

