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
   FwuHeciMsgs.h
Abstract:
   Definitions for the HECI Msgs for Fw Update Kernel Service.
Author
   Shivashankari Janakiraman
*/


#ifndef _FWU_COMMON_H
#define  _FWU_COMMON_H

#include "ver.h"
#include "FwCapsMsgs.h"

typedef UINT32 STATUS;
#define BLACK_LIST_ENTRY_MAX           10
#define FWU_PWD_MAX_SIZE               32
#define MAXIMUM_IPU_SUPPORTED          4
typedef enum
{
   FWU_FULL_UPDATE_OPERATION   =   0,
   FWU_IPU_UPDATE_OPERATION
}FWU_OPERATION;




typedef enum
{
   FWU_ENV_MANUFACTURING = 0,   // Manufacturing update
   FWU_ENV_IFU,                 // Independent Firmware update
}FWU_ENVIRONMENT;



typedef enum
{
   FWU_ROLLBACK_NONE = 0,       // No rollback
   FWU_ROLLBACK_1,              // Rollback 1 procedure
   FWU_ROLLBACK_2,              // Rollback 2 procedure
}FWU_ROLLBACK_MODE;




/*
typedef enum
{
   FWU_HOST_RESET_REQUIRED = 0,
   FWU_ME_RESET_REQUIRED,
   FWU_GLOBAL_RESET_REQUIRED
}FWU_HOST_RESET_TYPE;
*/

// Typedef for the commands serviced by the Fw Update service
typedef enum 
{
   FWU_GET_VERSION = 0,
   FWU_GET_VERSION_REPLY,
   FWU_START,
   //FWU_START2,
   FWU_START_REPLY,
   FWU_DATA,
   FWU_DATA_REPLY,
   FWU_END,
   FWU_END_REPLY,
   FWU_GET_INFO,
   FWU_GET_INFO_REPLY,
   FWU_GET_FEATURE_STATE,
   FWU_GET_FEATURE_STATE_REPLY,
   FWU_GET_FEATURE_CAPABILITY,
   FWU_GET_FEATURE_CAPABILITY_REPLY,
   FWU_GET_PLATFORM_TYPE,
   FWU_GET_PLATFORM_TYPE_REPLY,
   FWU_VERIFY_OEMID,
   FWU_VERIFY_OEMID_REPLY,
   FWU_GET_OEMID,
   FWU_GET_OEMID_REPLY,
   FWU_IMAGE_COMPATABILITY_CHECK,
   FWU_IMAGE_COMPATABILITY_CHECK_REPLY,
   FWU_GET_UPDATE_DATA_EXTENSION,
   FWU_GET_UPDATE_DATA_EXTENSION_REPLY,
   FWU_GET_RESTORE_POINT_IMAGE,
   FWU_GET_RESTORE_POINT_IMAGE_REPLY,
   //FWU_GET_RECOVERY_MODE,
   //FWU_GET_RECOVERY_MODE_REPLY,
   FWU_GET_IPU_PT_ATTRB,
   FWU_GET_IPU_PT_ATTRB_REPLY,
   FWU_GET_FWU_INFO_STATUS,
   FWU_GET_FWU_INFO_STATUS_REPLY,
   GET_ME_FWU_INFO,
   GET_ME_FWU_INFO_REPLY,

   FWU_INVALID_REPLY = 0xFF
} FWU_HECI_MESSAGE_TYPE;

#pragma pack(1)

typedef struct _FWU_GET_VERSION_MSG_REPLY
{
   UINT32      MessageType;
   UINT32      Status;
   UINT32      Sku;       
   UINT32      PCHVer;
   UINT32      Vendor;
   UINT32      LastFwUpdateStatus;
   UINT32      HwSku;
   VERSION     CodeVersion;
   VERSION     AMTVersion;
   UINT16      EnabledUpdateInterfaces;   // local, remote (LMS/LME) and secure update
   UINT16      SvnInFlash;                // Security version of image that is already in flash
   UINT32      DataFormatVersion;         // Upper 16 sig bits for Major version lower 16's for Minor Version
   UINT32      LastUpdateResetType;       // Last successfull update partition reset type prior to reboot. After reboot, it should be zero
} FWU_GET_VERSION_MSG_REPLY;

#define BIOS_BOOT_STATE_PRE_BOOT    0
#define BIOS_BOOT_STATE_POST_BOOT   2

typedef struct _FWU_GET_INFO_MSG
{
   UINT32 MessageType;
} FWU_GET_INFO_MSG;

// Contains the data to be returned for GET_VERSION command
typedef struct _FWU_GET_INFO_MSG_REPLY{
   UINT32          MessageType;
   UINT32          Status;
   VERSION         MEBxVersion;
   UINT32          FlashOverridePolicy;
   UINT32          MangeabilityMode;
   UINT32          BiosBootState;
   struct {
     UINT32        CryptoFuse   :1;
     UINT32        FlashProtection:1; // read from SPI driver
     UINT32        Obsolete_FwOverrideQualifier:2;
     UINT32        MeResetReason:2; // Reset.h
     UINT32        Obsolete_FwOverrideCounter:8; //TO DO: change this in MeTypes.h to UINT8
     UINT32        reserved:18;
    }Fields;
   UINT8          BiosVersion[20];

}FWU_GET_INFO_MSG_REPLY;

typedef struct
{
   unsigned long Data1;
   unsigned short Data2;
   unsigned short Data3;
   unsigned char Data4 [ 8 ];
} OEM_UUID;

typedef struct _FWU_GET_FEATURE_STATE_MSG_REPLY
{
   UINT32      MessageType;
   UINT32      Status;
   UINT32      FeatureState;       
} FWU_GET_FEATURE_STATE_MSG_REPLY;

typedef struct _FWU_GET_PLATFORM_TYPE_MSG_REPLY
{
   UINT32      MessageType;
   UINT32      Status;
   UINT32      PlatformType;       
} FWU_GET_PLATFORM_TYPE_MSG_REPLY;

typedef struct _FWU_VERIFY_OEMID_MSG
{
   UINT32  MessageType;
   OEM_UUID      OemId;
} FWU_VERIFY_OEMID_MSG;

typedef struct _FWU_VERIFY_OEMID_MSG_REPLY
{
   UINT32  MessageType;
   UINT32 Status;
}FWU_VERIFY_OEMID_MSG_REPLY;

typedef struct _FWU_START_MSG
{
   UINT32   MessageType;     
   UINT32   Length;           // Length of update image
   UINT8    UpdateType;       // 0 Full update, 1 partial IPU pdate
   UINT8    PassWordLength;   // Length of password not include NULL
   UINT8    PassWordData[FWU_PWD_MAX_SIZE];  // Password data not include NULL byte
   UINT32   IpuIdTobeUpdated; // Only for Partial FWU
   UINT32   UpdateEnvironment;// 0 default to normal manufacturing use 
                              // 1 is for Emergency IFU update 
   UINT32	UpdateFlags;	  // Currently only bit 0 is used to signify Restore Point 
   OEM_UUID OemId;
   UINT32   Resv[4];
} FWU_START_MSG;


typedef struct _FWU_START_MSG_REPLY
{
   UINT32   MessageType;
   STATUS   Status;
   UINT32   Resv[4];

}FWU_START_MSG_REPLY;



typedef struct _FWU_FLASH_IMAGE_START_MSG_REPLY
{
   STATUS   Status;
   UINT32   Resv[4];
}FWU_FLASH_IMAGE_START_MSG_REPLY;



typedef struct _FWU_IMAGE_COMPATABILITY_CHECK_MSG
{
   UINT32   MessageType;
   UINT32   ManifestLength;      // Manifest Len
   UINT8    Reserved[3];
   UINT8    ManifestBuffer[1];   // At least one element 
                                 // otherwise compiling error when use WATCOM compiler
}FWU_IMAGE_COMPATABILITY_CHECK_MSG;



typedef struct _FWU_IMAGE_COMPATABILITY_CHECK_MSG_REPLY
{
   UINT32  MessageType;
   STATUS      Status;  // 0 is OK for update, else failures
                        // Possible error code 
                        // STATUS_UPDATE_READ_FILE_FAILURE
                        // STATUS_UPDATE_FW_VERSION_MISMATCH
                        // STATUS_UPDATE_IMAGE_INVALID
                        // STATUS_UPDATE_FLASH_CODE_PARTITION_INVALID
                        // STATUS_UPDATE_IMAGE_VERSION_HISTORY_CHECK_FAILURE
                        // STATUS_UPDATE_IMAGE_BLACKLISTED
}FWU_IMAGE_COMPATABILITY_CHECK_MSG_REPLY;


typedef struct    _BLACK_LIST_ENTRY
{
   UINT16   ExpressionType;
   UINT16   MinorVer;
   UINT16   HotfixVer1;
   UINT16   BuildVer1;
   UINT16   HotfixVer2;
   UINT16   BuildVer2;
}BLACK_LIST_ENTRY;


typedef struct _UPDATE_VERSION_INFO
{
   VERSION  Version;
   UINT8    History[4];
   UINT32   CriticalHotfixDescriptor;

}UPDATE_VERSION_INFO;


typedef struct 
{
    UINT32 MessageType;
} FWU_GET_UPDATE_DATA_EXTENSION_MSG;

typedef struct _FWU_GET_UPDATE_DATA_EXTENSION_MSG_REPLY
{
   UINT32       MessageType;
   STATUS       Status;          // 0 for success, other failure and the info is invalid
   UINT8        History[4];      // Minor0Predecessor, Minor1Predecessor, Minor2Predecessor, Minor3Predecessor
   UINT32       CriticalHotfixDescriptor;
   BLACK_LIST_ENTRY     BlackList[BLACK_LIST_ENTRY_MAX];
}FWU_GET_UPDATE_DATA_EXTENSION_MSG_REPLY;


typedef struct _FWU_DATA_MSG
{
   UINT32   MessageType;
   UINT32   Length;
   UINT8    Reserved[3];
   UINT8    Data[1];
} FWU_DATA_MSG;


typedef struct _FWU_DATA_MSG_REPLY
{
   UINT32  MessageType;
   STATUS Status;
}FWU_DATA_MSG_REPLY;


typedef struct _FWU_GET_FWU_INFO_STATUS_MSG
{
   UINT32   MessageType;
   UINT32   InfoParm;         // Not used 
   UINT32   Resv[4];
}FWU_GET_FWU_INFO_STATUS_MSG;



typedef struct _FWU_INFO_FLAGS
{
   UINT32 RecoveryMode:2;   // 0 = No recovery; 1 = Full Recovery Mode,2 = Partial Recovery Mode (unused at present)
   UINT32 IpuNeeded:1;      // IPU_NEEDED bit, if set we are in IPU_NEEDED state.
   UINT32 FwInitDone:1;     // If set indicate FW is done initialized
   UINT32 FwuInProgress:1;  // If set FWU is in progress, this will be set for IFU update as well
   UINT32 SuInprogress:1;   // If set IFU Safe FW update is in progress. 
   UINT32 NewFtTestS:1;     // If set indicate that the new FT image is in Test Needed state (Stage 2 Boot)
   UINT32 SafeBootCnt:4;    // Boot count before the operation is success
   UINT32 FsbFlag:1;        // Force Safe Boot Flag, when this bit is set, we'll boot kernel only and go into recovery mode	

   //////////////////////////////////////////////////////
   // These fields below are important for FWU tool. 
   //////////////////////////////////////////////////////
   UINT32 LivePingNeeded:1;     // Use for IFU only, See Below  
                                // FWU tool needs to send Live-Ping or perform querying to confirm update successful.
                                // With the current implementation when LivePingNeeded is set, 
                                // Kernel had already confirmed it. No action from the tool is needed.
   UINT32 ResumeUpdateNeeded:1; // Use for IFU only, If set FWU tool needs to resend update image
   UINT32 RollbackNeededMode:2; // FWU_ROLLBACK_NONE = 0, FWU_ROLLBACK_1, FWU_ROLLBACK_2 
                                // If not FWU_ROLLBACK_NONE, FWU tool needs to send restore_point image. 
   UINT32 ResetNeeded:2;        // When this field is set to ME_RESET_REQUIRED, FW Kernel will
                                // perform ME_RESET after this message. No action from the tool is needed.
   UINT32 Reserve:14;
}FWU_INFO_FLAGS;

// LivePingNeeded		ResumeUpdateNeeded		    Stage
// 0						0							1
// 0						1							2
// 1						0							3


typedef struct _FWU_GET_FWU_INFO_STATUS_MSG_REPLY
{
   UINT32   MessageType;
   STATUS   Status;
   FWU_INFO_FLAGS   Flags;
   UINT32   Resv[4];
}FWU_GET_FWU_INFO_STATUS_MSG_REPLY;



typedef struct _PT_ATTRB
{
   UINT32   PtNameId;      // HW_COMP_HDR_STRUCTID_WCOD     0x244f4357 OR 
                           // HW_COMP_HDR_STRUCTID_LOCL     0x4C434F4C OR
                           // HW_COMP_HDR_STRUCTID_MDMV     0x564D444D 
   UINT32   LoadAddress;         // Load Address of the IPU
   VERSION  FwVer;         // FW version from IUP Manifest
   UINT32   CurrentInstId; // Current Inst ID from flash, 0 indicate invalid ID 
   UINT32   CurrentUpvVer; // Upper sig 16 bits are Major Version.
   UINT32   ExpectedInstId;// Expected Inst ID that need to be updated to
   UINT32   ExpectedUpvVer;// Upper sig 16 bits are Major Version.
   UINT32   Resv[4];
}PT_ATTRB;


typedef struct _FWU_GET_IPU_PT_ATTRB_MSG
{
   UINT32   MessageType;
} FWU_GET_IPU_PT_ATTRB_MSG;


typedef struct _FWU_GET_IPU_PT_ATTRB_MSG_REPLY
{
   UINT32   MessageType;      // Internal FWU tool use only
   STATUS   Status;           // Internal FWU tool use only
   VERSION  FtpFwVer;         // FW version in Fault Tolerance Partition. 
                              // This might be used for diagnostic or debug.
   UINT32   SizeoOfPtAttrib;  // Size in bytes. Simply is the sizeof (PT_ATTRB structure)
   UINT32   NumOfPartition;   // Number of partition actually return in this reply message
   PT_ATTRB  PtAttribute[MAXIMUM_IPU_SUPPORTED];
   UINT32   Resv[4];
}FWU_GET_IPU_PT_ATTRB_MSG_REPLY;



/*
typedef struct _FWU_GET_RECOVERY_MODE_MSG
{
   UINT32   MessageType;
}FWU_GET_RECOVERY_MODE_MSG;


typedef struct _FWU_GET_RECOVERY_MODE_MSG_REPLY
{
   UINT32   MessageType;
   STATUS   Status;
   UINT32   ReoveryMode; // 0 No recovery, 
                         // 1 NFTP is invalid, Full FW update is required
                         // 2 Only one or more IPU Partition is invalid.
                         
}FWU_GET_RECOVERY_MODE_MSG_REPLY;
*/



/**
 * FWU_END_MESSAGE - end the update process
 *
 * @MessageType: FWU_MESSAGE_TYPE_END
 */
typedef struct {
    UINT32 MessageType;
} FWU_END_MESSAGE;

typedef struct _FWU_END_MSG_REPLY
{
   UINT32  MessageType;
   STATUS Status; // 0 indicate success, else failure
   UINT32 ResetType; 
   UINT32 Resv[4];

}FWU_END_MSG_REPLY;

typedef struct _FWU_GET_OEMID_MSG_REPLY
{
   UINT32  MessageType;
   STATUS        Status;
   OEM_UUID      OemId;
} FWU_GET_OEMID_MSG_REPLY;

typedef struct 
{
    UINT32 MessageType;
} FWU_GET_RESTORE_POINT_IMAGE_MESSAGE;

typedef struct _FWU_GET_RESTORE_POINT_IMAGE_MSG_REPLY
{
   UINT32  MessageType;
   STATUS  Status;
   UINT32  RestorePointImageSize; // Size of image is in Bytes
   UINT32  RestorePointImage[1];
}FWU_GET_RESTORE_POINT_IMAGE_MSG_REPLY;


typedef struct _FWU_INVALID_MSG_REPLY
{
   UINT32  MessageType;
   STATUS Status;
}FWU_INVALID_MSG_REPLY;


//Data
typedef struct
{
   UINT32   Length;
   UINT8    Reserved[3];
   UINT8    Data[1];
} FWU_HECI_MESSAGE_DATA;

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////// ME FWU related Information that will be provided by discovery DLL /////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
 
// ME_FWU_INFO_MSG General information related to FW update
// Few basic rules:
// Do not delete member
// Allow add new member only to the end.
// The producer(Kernel) will only copy up to the size of the consumer StructSize. Consumer must set the StructSize.
// TBD each field need to initialize to invalid pattern? In the case that we could not get 
// the information. 
typedef struct  _ME_FWU_INFO
{
   UINT32      StructSize;          // SizeOfThisStruct
   UINT32      ApiVer;              // Version of this API
   VERSION     FtpVer;              // QWORD Major(15:0), Minor(31:16), HotFix(47:32), 
                                    // Build#(63:48)
   VERSION     NftpVer;             // QWORD Major(15:0), Minor(31:16), HotFix(47:32), 
                                    // Build#(63:48)
   UINT32      ChipsetVer;          // PCH version
   UINT32      GlobalChipId;        // Global chip identification
   UINT8	   SystemManufacturer[32]; // Ascii of char string 

   UINT32      MebxFwuConfig;       // 0= Disable, 1 = enable , 2 = PW protected
   MEFWCAPS_SKU      HwSku;				// MEFWCAPS_SKU   HW feature set or not
   MEFWCAPS_SKU      FwSku;				// MEFWCAPS_SKU   FW feature set from 
                                    // SkuTable fuse or not fuse  
   UINT32      LastFwUpdateStatus;  // Last FW update status
   UINT32      DataFormatVer;       // Data format version Major(31:16), Minor (15:0). Only Major is used
   UINT32      SvnVer;              // Security version: Major (31:16), Minor (15:0),Only Major is used
   UINT32      VcnVer;              // Version Control Number: Major (31:16), Minor (15:0),Only Major is used
                                    // This field is currently not used
   VERSION     MebxVer;             // MEBX version 
   FWU_INFO_FLAGS  FwuInfoFlags;     // FWU information flags
   ME_PLATFORM_TYPE  PlatformAttributes;	// ME_PLATFORM_TYPE Contains Platform Attributes: 
                                            // CpuType,platform type, superSku,etc. See ME_PLATFORM_TYPE
   OEM_UUID    OemId;               // Famous OEM_ID
   UINT16      MeFwSize;            // Size of FW image in multiple of .5MB
   UINT8       History[4];          // Keep track of version tree history
                                    // Minor0Predecessor, Minor1Predecessor, Minor2Predecessor, Minor3Predecessor
                                    // FWU will check to see if the update image has the same predecessor with the
                                    // one that is already in the flash before allow the update.
   UINT32      CriticalHotfixDescriptor;  // Each bit signify a particular Critical Hot Fix
   BLACK_LIST_ENTRY       BlackListEntry[BLACK_LIST_ENTRY_MAX]; // Black list entries. Note that it might go away in 8.0
   UINT16      NumSupportedIup;     // Number of IPUs actually supported, use this number to traverse IupEntry  
   PT_ATTRB    IupEntry [MAXIMUM_IPU_SUPPORTED];  

}ME_FWU_INFO;



typedef struct _ME_FWU_INFO_MSG_REPLY
{
	UINT32				MessageType;
	UINT32				Status; // 0 = success, else failure
	ME_FWU_INFO			MeFwuInfo;
}ME_FWU_INFO_MSG_REPLY;

typedef struct _ME_FWU_INFO_MSG
{
	UINT32 MessageType;
	UINT32 MessageParams[2]; //Currently not used
}ME_FWU_INFO_MSG;


//NOTE: WE ARE VERY CLOSE TO MAXIMUM HECI MESSAGE BUFFER SIZE
typedef struct  _FWU_HECI_MSG
{
   union{
      UINT32  MessageType;
 
      FWU_GET_VERSION_MSG_REPLY        VersionReply;
      FWU_START_MSG                    Start;
      FWU_START_MSG_REPLY              StartReply;
      FWU_DATA_MSG                     Data;
      FWU_DATA_MSG_REPLY               DataReply;
      FWU_END_MSG_REPLY                EndReply;
      FWU_GET_INFO_MSG_REPLY           InfoReply;
      FWU_INVALID_MSG_REPLY            InvalidMsgReply;
      FWU_GET_FEATURE_STATE_MSG_REPLY  FeatureStateReply;
      FWU_GET_PLATFORM_TYPE_MSG_REPLY  PlatformTypeReply;
      FWU_VERIFY_OEMID_MSG             VerifyOemId;
      FWU_VERIFY_OEMID_MSG_REPLY       VerifyOemIdReply;
      FWU_GET_OEMID_MSG_REPLY          GetOemIdReply;
      FWU_IMAGE_COMPATABILITY_CHECK_MSG   ImageCheck;
      FWU_IMAGE_COMPATABILITY_CHECK_MSG_REPLY         ImageCheckReply;
      FWU_GET_UPDATE_DATA_EXTENSION_MSG_REPLY         GetUpdateDataExtReply;
      FWU_GET_RESTORE_POINT_IMAGE_MSG_REPLY           GetRestorePointImageReply;
      //FWU_GET_RECOVERY_MODE_MSG                       GetRecoveryMode; 
      //FWU_GET_RECOVERY_MODE_MSG_REPLY                 GetRecoveryModeReply; 
      FWU_GET_IPU_PT_ATTRB_MSG                        GetIpuPtAttrb; 
      FWU_GET_IPU_PT_ATTRB_MSG_REPLY                  GetIpuPtAttrbReply;
      FWU_GET_FWU_INFO_STATUS_MSG                     GetFwuInfoStatusMsg;
      FWU_GET_FWU_INFO_STATUS_MSG_REPLY               GetFwuInfoStatusMsgReply;
	  ME_FWU_INFO_MSG_REPLY					GetMeInfoMsgReply;

   }MessageData;
}FWU_HECI_MSG;


//FWU_STATUS_INVALID_ACCESS
// Error messages sent to HECI tool

typedef enum _FWU_HECI_MSG_STATUS{
      FWU_NOT_READY,
      FWU_ILLEGAL_LENGTH, //allocate image buffer
} FWU_HECI_MSG_STATUS;



typedef enum
{
   BLE_EMPTY = 0,
   BLE_EQ,
   BLE_LTE,
   BLE_GTE,
   BLE_RANGE
}BLACK_LIST_EXPRESSION_TYPES;



#pragma pack()

#endif  // _FWU_COMMON_H
