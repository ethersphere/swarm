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
   FwCapsMsgs.h
Abstract:
   Contains data structures and constants used in ME Configuration and application Rules
Authors:
   Tam Nguyen
*/

#ifndef _FW_CAPS_MSGS_H
#define _FW_CAPS_MSGS_H

#include "MkhiMsgs.h"

#pragma warning (disable: 4214 4200)
#pragma pack(1)

#ifndef DEFAULT_AMT_WOL_TIMEOUT_MINUTES
// default value for AMT WOL Timer (65535 min)
#define DEFAULT_AMT_WOL_TIMEOUT_MINUTES 65535
#endif


//Host configure commands
#define FWCAPS_GET_RULE_CMD            0x02
#define FWCAPS_GET_RULE_CMD_ACK        0x82
#define FWCAPS_SET_RULE_CMD            0x03
#define FWCAPS_SET_RULE_CMD_ACK        0x83
#define FWCAPS_GET_RULE_STATE_CMD      0x05
#define FWCAPS_GET_RULE_STATE_CMD_ACK  0x85

#define FWCAPS_RULE_LOCKED            BIT1
#define FWCAPS_RULE_LOCKABLE          BIT2
#define FWCAPS_RULE_EXTERNAL          BIT3
#define FWCAPS_RULE_POST_PRODUCTION   BIT4

#define FWCAPS_RULE_STATE_CLEARED    0


#define SIZE_OF_HDR_AND_RULE_ID ((sizeof(MKHI_MESSAGE_HEADER)) + (sizeof(RULE_ID)))
//SIZE_OF_HDR_AND_RULE_ID + size of RuleLength
#define GET_ACK_RULE_DATA_START_POS ((SIZE_OF_HDR_AND_RULE_ID) + sizeof(UINT8)) 
//
//ME Configuration Manager Rule ID definition. Note that RuleTypeId is used as
//index into rules table, so it must be sequencial (e.g. first RuleTypeId is 0, 
//second RuleTypeId is 1, ...). Sku indicates application sku and AppRules 
//indicates whether a rule is application of ME configuration rule.
//

//Arbitrary max number of rules per app
#define FWCAPS_APP_RULES_MAX       20
//rule size for Bios tables fingerprints is 259 bytes
#define FWCAPS_RULE_SIZE_MAX      (260)

#define FWCAPS_APPS_MAX           5 


/** @brief  _ME_FEATURE_ID The following enumeration defines the various features supported by ME.
  * These feature IDs can be used to determine the state of a particular feature.
  * @deprecated ME_FID_CRYPTO No longer supported and always disabled.
  *
*/

typedef enum _ME_FEATURE_ID 
{
   /**
 * @brief
 */
   /*ME_FID_AMT = 0,*/   // This is AMT CORP w/ VPRO or SoftCreek upgrade 
   ME_FID_MNG_FULL = 0,
/**
 * @brief
 */
   /*ME_FID_AMT_STD,*/
   ME_FID_MNG_STD,
/**
 * @brief
 */
   /*ME_FID_AMT_CONS,*/
   ME_FID_AMT,
/**
 * @brief
 */
   ME_FID_LOCAL_MNG,
/**
 * @brief
 */
   ME_FID_L3_MNG, // Not included in CPT SKU matrix
/**
 * @brief
 */   
   ME_FID_TDT,
/**
 * @brief
 */
   ME_FID_SOFTCREEK,
/**
 * @brief
 */
   ME_FID_VE,
/**
 * @brief
 */
   ME_FID_NAND35,
/**
 * @brief
 */
   ME_FID_NAND29,
/**
 * @brief
 */
   ME_FID_THERM_REPORT,
/**
 * @brief
 */
   ME_FID_ICC_OVERCLOCK,
/**
 * @brief
 */
   ME_FID_PAV,
/**
 * @brief
 */
   ME_FID_SPK,
/**
 * @brief
 */
   ME_FID_RCA,
   /**
 * @brief
 */
   ME_FID_RPAT,
   /**
 * @brief
 */
   ME_FID_HAP,
/**
 * @brief
 */
   ME_FID_IPV6,
/**
 * @brief
 */
   ME_FID_KVM, 
   /**
 * @brief
 */
   ME_FID_OCH,
   /**
 * @brief
 */
   ME_FID_MEDAL,
/**
 * @brief
 */
  ME_FID_TLS_CONF,
/**
 * @brief
 */
   ME_FID_CILA, 
 /**
 * @brief
 */
   ME_FID_WLAN, 
/**
 * @brief
 */
   ME_FID_WL_DISP, 
/**
 * @brief
 */
   ME_FID_USB3, 
 /**
 * @brief
 */
   ME_FID_NAP,
/**
 * @brief
 */
 ME_FID_ALARMCLK,
 /**
 * @brief
 */
 ME_FID_CBRAID,
 /**
 * @brief
 */
 ME_FID_MEDIAVAULT,
  /**
 * @brief
 */
 ME_FID_MDNSPROXY,
  /**
 * @brief
 */
   ME_FID_MAX = 32,
/**
 * @brief this is for BIST manager exclusively.  
 */
   ME_FID_uKERNEL,
/**
 * @brief this is for BIST manager exclusively.  
 */
   ME_FID_POLICY,  
/**
 * @brief this is for BIST manager exclusively.  
 */
   ME_FID_COMMON_SERVICES,
/**
 * @brief this is for BIST manager exclusively.  
 */
   ME_FID_MCTP,  
/**
 * @brief Last item for BIST manager.  
 */
   ME_FID_BIST_MAX
} ME_FEATURE_ID;


/**
*
* @brief RULE_ID
*
*/

typedef union _RULE_ID
{
   UINT32      Data;
   struct
   {
      UINT32   RuleTypeId     :16;
      UINT32   FeatureId      :8;
      UINT32   Reserved       :8;
   }Fields;
}RULE_ID;
C_ASSERT(sizeof(RULE_ID) == 4);

/**
*
* @brief RULE_CFG_INFO
*
*/
typedef union _RULE_CFG_INFO
//Rule Attributes. Control attributes of a rule
{
   UINT8     Data;
   struct 
   {
      UINT8  Reserved             :1;
      UINT8  Locked               :1;
      UINT8  Lockable             :1;
      UINT8  ExternallyUpdateable :1;
      UINT8  PostProduction       :1;
      UINT8  Reserved2            :3;
   }Fields;
}RULE_CFG_INFO;
C_ASSERT(sizeof(RULE_CFG_INFO) == 1);

//This is the definition of ME Configuration Rule.
typedef struct _FWCAPS_RULE
{
   RULE_ID         RuleId;
   RULE_CFG_INFO   RuleCfgInfo;
   UINT8           Reserved[2];  
   UINT8           Size;
   UINT32          Data;
}FWCAPS_RULE;
C_ASSERT(sizeof(FWCAPS_RULE) == 12);

//HECI message get data structure. This is the message structure sent from HCI.
typedef struct _GET_RULE_DATA
{
   RULE_ID  RuleId;
}GET_RULE_DATA;

typedef struct _FWCAPS_GET_RULE
{
   MKHI_MESSAGE_HEADER     Header;
   GET_RULE_DATA           Data;
}FWCAPS_GET_RULE;

#ifndef __WATCOM__
//HECI message get response data structure. This is the message sent from
//ME Configuration Manager
typedef struct _GET_RULE_ACK_DATA
{
   RULE_ID  RuleId;
   UINT8    RuleDataLen;
   UINT8    RuleData[0];
}GET_RULE_ACK_DATA;

typedef struct _FWCAPS_GET_RULE_ACK
{
   MKHI_MESSAGE_HEADER     Header;
   GET_RULE_ACK_DATA       Data;
}FWCAPS_GET_RULE_ACK;

//HECI message set data structure. This is the message structure sent from HCI.

typedef struct _SET_RULE_DATA
{
   RULE_ID  RuleId;
   UINT8    RuleDataLen;
   UINT8    RuleData[0];
}SET_RULE_DATA;

typedef struct _FWCAPS_SET_RULE
{
   MKHI_MESSAGE_HEADER     Header;
   SET_RULE_DATA           Data;
}FWCAPS_SET_RULE;
#endif // __WATCOM__

//HECI message set response data structure. This is the message sent from
//ME Configuration Manager
typedef struct _SET_RULE_ACK_DATA
{
   RULE_ID  RuleId;
}SET_RULE_ACK_DATA;

typedef struct _FWCAPS_SET_RULE_ACK
{
   MKHI_MESSAGE_HEADER     Header;
   SET_RULE_ACK_DATA       Data;
}FWCAPS_SET_RULE_ACK;

//HECI message set data structure. This is the message structure sent from HCI.
typedef struct _GET_RULE_STATE_DATA
{
   RULE_ID  RuleId;
}GET_RULE_STATE_DATA;

typedef struct _FWCAPS_GET_RULE_STATE
{
   MKHI_MESSAGE_HEADER     Header;
   GET_RULE_STATE_DATA     Data;
}FWCAPS_GET_RULE_STATE;

//HECI message get metadata response data structure. This is the message sent 
//from ME Configuration Manager
typedef struct _GET_RULE_STATE_ACK_DATA
{
   RULE_ID        RuleId;
   RULE_CFG_INFO  RuleMetaData;
}GET_RULE_STATE_ACK_DATA;

typedef struct _FWCAPS_GET_RULE_STATE_ACK
{
   MKHI_MESSAGE_HEADER        Header;
   GET_RULE_STATE_ACK_DATA    Data;
}FWCAPS_GET_RULE_STATE_ACK;

//Macro to build a rule identifier. for Me rules all other fields are zeros
#define MAKE_ME_RULE_ID(FeatureId, RuleId)  ((FeatureId << 16) | RuleId)

//ME Configuration rule ID Type
#define MEFWCAPS_FW_SKU_RULE                     0
#define MEFWCAPS_MANAGEABILITY_SUPP_RULE         1
#define MEFWCAPS_QST_STATE_RULE                  2
#define MEFWCAPS_CB_STATE_RULE                   3
#define MEFWCAPS_LAN_STATE_RULE                  4
#define MEFWCAPS_LAN_SKU_RULE                    5
#define MEFWCAPS_ME_PLATFORM_STATE_RULE          6
#define MEFWCAPS_ME_LOCAL_FW_UPDATE_RULE         7
#define MEFWCAPS_TLS_CONF_STATE_RULE             8
// #define MEFWCAPS_LOCAL_FW_UPD_OVR_CNTR_RULE   9    // Deprecated
// #define MEFWCAPS_LOCAL_FW_UPD_OVR_QUAL_RULE   10   // Deprecated
// #define MEFWCAPS_FDOPS_USAGE_RULE             11   // Deprecated
#define MEFWCAPS_OEM_SKU_RULE                    12
#define MEFWCAPS_LAN_BLOCK_TRAFFIC_RULE          13
#define MEFWCAPS_DT_RULE                         14
// Rules for Platform Configuration
#define MEFWCAPS_PCV_LAN_WELL_CONFIG_RULE             15
#define MEFWCAPS_PCV_WLAN_WELL_CONFIG_RULE            16
#define MEFWCAPS_PCV_CPU_MISSING_LOGIC_RULE           17
#define MEFWCAPS_PCV_M3_POWER_RAILS_PRESENT_RULE      18 
#define MEFWCAPS_PCV_ICC_OEM_LAYOUT_RULE              19
#define MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO1_RULE         20
#define MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO2_RULE         21
#define MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO3_RULE         22 
#define MEFWCAPS_PCV_ICC_ME_EC_SPEC_COMPLIANT_RULE    23
#define MEFWCAPS_PCV_ICC_FPS_PWR_CTRL_MGPIO_RULE      24
#define MEFWCAPS_PCV_ICC_FPS_INTERRUPT_MGPIO_RULE     25
#define MEFWCAPS_PCV_ICC_THERM_MON_MGPIO_RULE         26
#define MEFWCAPS_PCV_DOCK_IND_MGPIO_RULE              27
#define MEFWCAPS_PCV_OEM_CAP_CFG_RULE                 28
#define MEFWCAPS_PCV_OEM_PLAT_TYPE_CFG_RULE           29
#define MEFWCAPS_PCV_SUS_WELL_DOWN_S45_MOFF_DC_RULE   30 
#define MEFWCAPS_FOV_MANUF_STATUS_RULE                31
#define MEFWCAPS_FEATURE_ENABLE_RULE                  32
#define MEFWCAPS_STATE_FOR_ALL_FEATURES_RULE          33
#define MEFWCAPS_CHECK_OEM_CAPS_RULE                  34
#define MEFWCAPS_CHECK_USER_CAPS_RULE                 35
#define MEFWCAPS_FEATURE_ACTIVE_RULE                  36
#define MEFWCAPS_PCV_TARGET_MARKET_TYPE_CFG_RULE      37   
#define MEFWCAPS_PCV_ENABLE_CLINK_RULE                38
#define MEFWCAPS_AVAILABLE_BITMAP_RULE                39
#define MEFWCAPS_CPU_STR_EMULATION_RULE               40
#define MEFWCAPS_PCV_ENABLE_MOFFOVERRIDE_RULE         41
#define MEFWCAPS_QMQS_TO_HM_CONV_RULE                 42
#define MEFWCAPS_OEM_TAG_RULE                         43
#define MEFWCAPS_IPU_NEEDED_STATE_RULE                44 
// C-Link override rule is different from C-Link disable/enable rule
// and was added to maintain backward compatibility. If C-Link override 
// is put on, C-Link cannot be enabled via any option (strap, Nvar or MKHI mgs)
// Override needs to be lifted off for C-Link to be enabled via any option.
#define MEFWCAPS_CLINK_OVERRIDE_RULE                  45
#define MEFWCAPS_ME_FWU_IFR_RULE                      46   // 0 IFR Not Allowed;  1 IFR Allowed
#define MEFWCAPS_MAX_RULES                            47
#define MEFWCAPS_INVALID_RULE                         999

#define ME_RULE_FEATURE_ID                       0
//ME Configuration rule ID
#define MEFWCAPS_SKU_RULE_ID                     MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_FW_SKU_RULE)
#define MEFWCAPS_MANAGEABILITY_SUPP_RULE_ID      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_MANAGEABILITY_SUPP_RULE)
#define MEFWCAPS_QST_STATE_RULE_ID               MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_QST_STATE_RULE)
#define MEFWCAPS_CB_STATE_RULE_ID                MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_CB_STATE_RULE)
#define MEFWCAPS_LAN_STATE_RULE_ID               MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_LAN_STATE_RULE)
#define MEFWCAPS_LAN_SKU_RULE_ID                 MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_LAN_SKU_RULE)
#define MEFWCAPS_ME_PLATFORM_STATE_RULE_ID       MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_ME_PLATFORM_STATE_RULE)
#define MEFWCAPS_ME_LOCAL_FW_UPDATE_RULE_ID      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_ME_LOCAL_FW_UPDATE_RULE)
#define MEFWCAPS_TLS_CONF_STATE_RULE_ID          MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_TLS_CONF_STATE_RULE)
//#define MEFWCAPS_LOCAL_FW_UPD_OVR_CNTR_RULE_ID   MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_LOCAL_FW_UPD_OVR_CNTR_RULE)  
//#define MEFWCAPS_LOCAL_FW_UPD_OVR_QUAL_RULE_ID   MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_LOCAL_FW_UPD_OVR_QUAL_RULE)                                                                      
#define MEFWCAPS_OEM_SKU_RULE_ID                 MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_OEM_SKU_RULE)                                                                      
#define MEFWCAPS_LAN_BLOCK_TRAFFIC_RULE_ID       MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_LAN_BLOCK_TRAFFIC_RULE)
#define MEFWCAPS_DT_RULE_ID                      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_DT_RULE)  


// Rules from PRA
#define MEFWCAPS_PCV_LAN_WELL_CONFIG_RULE_ID         MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_LAN_WELL_CONFIG_RULE)
#define MEFWCAPS_PCV_WLAN_WELL_CONFIG_RULE_ID        MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_WLAN_WELL_CONFIG_RULE) 
#define MEFWCAPS_PCV_ENABLE_CLINK_RULE_ID               MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ENABLE_CLINK_RULE) 
#define MEFWCAPS_PCV_CPU_MISSING_LOGIC_RULE_ID       MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_CPU_MISSING_LOGIC_RULE) 
#define MEFWCAPS_PCV_M3_POWER_RAILS_PRESENT_RULE_ID   MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_M3_POWER_RAILS_PRESENT_RULE)   
#define MEFWCAPS_PCV_ICC_OEM_LAYOUT_RULE_ID          MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_OEM_LAYOUT_RULE)  
#define MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO1_RULE_ID      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO1_RULE)
#define MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO2_RULE_ID      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO2_RULE)
#define MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO3_RULE_ID      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_OEMRECSEL_GPIO3_RULE) 
#define MEFWCAPS_PCV_ICC_ME_EC_SPEC_COMPLIANT_RULE_ID  MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_ME_EC_SPEC_COMPLIANT_RULE)
#define MEFWCAPS_PCV_ICC_FPS_PWR_CTRL_MGPIO_RULE_ID    MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_FPS_PWR_CTRL_MGPIO_RULE)
#define MEFWCAPS_PCV_ICC_FPS_INTERRUPT_MGPIO_RULE_ID   MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_FPS_INTERRUPT_MGPIO_RULE)
#define MEFWCAPS_PCV_ICC_THERM_MON_MGPIO_RULE_ID       MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_ICC_THERM_MON_MGPIO_RULE)
#define MEFWCAPS_PCV_DOCK_IND_MGPIO_RULE_ID           MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_DOCK_IND_MGPIO_RULE)
#define MEFWCAPS_PCV_OEM_CAP_CFG_RULE_ID              MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_OEM_CAP_CFG_RULE)
#define MEFWCAPS_PCV_OEM_PLAT_TYPE_CFG_RULE_ID         MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_OEM_PLAT_TYPE_CFG_RULE)
#define MEFWCAPS_PCV_SUS_WELL_DOWN_S45_MOFF_DC_RULE_ID  MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_SUS_WELL_DOWN_S45_MOFF_DC_RULE)

// Added FOV manuf status as a kernel rule
#define MEFWCAPS_FOV_MANUF_STATUS_RULE_ID             MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_FOV_MANUF_STATUS_RULE)
#define MEFWCAPS_AVAILABLE_BITMAP_RULE_ID             MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_AVAILABLE_BITMAP_RULE)
#define MEFWCAPS_FEATURE_ENABLE_RULE_ID               MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_FEATURE_ENABLE_RULE)
#define MEFWCAPS_STATE_FOR_ALL_FEATURES_RULE_ID       MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_STATE_FOR_ALL_FEATURES_RULE )  
#define MEFWCAPS_CHECK_OEM_CAPS_RULE_ID               MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_CHECK_OEM_CAPS_RULE)
#define MEFWCAPS_CHECK_USER_CAPS_RULE_ID              MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_CHECK_USER_CAPS_RULE)    
#define MEFWCAPS_FEATURE_ACTIVE_RULE_ID               MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_FEATURE_ACTIVE_RULE)
#define MEFWCAPS_PCV_TARGET_MARKET_TYPE_CFG_RULE_ID   MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID,MEFWCAPS_PCV_TARGET_MARKET_TYPE_CFG_RULE)
#define MEFWCAPS_CPU_STR_EMULATION_RULE_ID            MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_CPU_STR_EMULATION_RULE)
#define MEFWCAPS_PCV_ENABLE_MOFFOVERRIDE_RULE_ID      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_PCV_ENABLE_MOFFOVERRIDE_RULE)
#define MEFWCAPS_QMQS_TO_HM_CONV_RULE_ID              MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_QMQS_TO_HM_CONV_RULE)
#define MEFWCAPS_OEM_TAG_RULE_ID                      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_OEM_TAG_RULE)
#define MEFWCAPS_IPU_NEEDED_STATE_RULE_ID             MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_IPU_NEEDED_STATE_RULE)
#define MEFWCAPS_CLINK_OVERRIDE_RULE_ID               MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_CLINK_OVERRIDE_RULE)
#define MEFWCAPS_ME_FWU_IFR_RULE_ID                   MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_ME_FWU_IFR_RULE)

#define MEFWCAPS_INVALID_RULE_ID                      MAKE_ME_RULE_ID(ME_RULE_FEATURE_ID, MEFWCAPS_INVALID_RULE )

// should use PLAT_TYPE message to get the mobile or desktop information
// instead of the following.
#define MEFWCAPS_MOBILE_PLATFORM_ENABLED     1
#define MEFWCAPS_DESKTOP_PLATFORM_ENABLED    2

#define MEFWCAPS_PLATFORM_IS_CORPORATE       1
#define MEFWCAPS_PLATFORM_IS_CONSUMER        2


//ME Configuration rules data structure

//to do: need to set the correct default...

//MEFWCAPS_FW_SKU
//    Indicates the firmware modules present in this SKU. This is an Intel
//    defined policy and it is not updateable by OEMs.
//    


/**
*
* @brief MEFWCAPS_SKU
*
*/
typedef union _MEFWCAPS_SKU
{
   UINT32   Data;
   struct
   {
      UINT32   MngFull          :1; // BIT 0:  Full network manageability 
      UINT32   MngStd           :1; // BIT 1:  Standard network manageability 
      UINT32   Amt              :1; // BIT 2:  Consumer manageability      
      UINT32   LocalMng             :1; // BIT 3:  Repurposed from IRWT, Local Mng a.k.a Treasurelake
      UINT32   L3Mng              :1; // BIT 4:  Repurposed from Qst
      UINT32   Tdt              :1; // BIT 5:  AT-p (Anti Theft PC Protection aka Tdt)
      UINT32   SoftCreek        :1; // BIT 6:  Intel Capability Licensing Service aka CLS
      UINT32   Ve               :1; // BIT 7:  Virtualization Engine
      UINT32   Nand35           :1; // BIT 8:  Tacoma Pass 35mm
      UINT32   Nand29           :1; // BIT 9:  Tacoma Pass 29mm
      UINT32   ThermReport      :1; // BIT 10: Thermal Reporting
      UINT32   IccOverClockin   :1; // BIT 11: 
      UINT32   Pav              :1; // BIT 12: Protected Audio Video Path (**Reserved for external documentation***)
      UINT32   Spk              :1; // BIT 13:
      UINT32   Rca              :1; // BIT 14:
      UINT32   Rpat             :1; // BIT 15:      
      UINT32   Hap              :1; // BIT 16: HAP_Platform
      UINT32   Ipv6             :1; // BIT 17:
      UINT32   Kvm              :1; // BIT 18: 
      UINT32   Och              :1; // BIT 19: 
      UINT32   MEDAL        :1; // BIT 20
      UINT32   Tls              :1; // BIT 21: 
      UINT32   Cila             :1; // BIT 22: 
      UINT32   Wlan             :1; // BIT 23: 
      UINT32   WirelessDisp     :1; // BIT 24: Wireless Display
      UINT32   USB3             :1; // BIT 25: USB 3.0
      UINT32   Nap              :1;  //BIT 26
      UINT32   AlarmClk         :1; //BIT 27
      UINT32   CbRaid              :1;//Bit 28
      UINT32   MediaVault          :1;//Bit 29
      UINT32   mDNSProxy          :1;//Bit 30
      UINT32   Nfc        :1; //Bit 31 NFC   
   }Fields;
}MEFWCAPS_SKU;
C_ASSERT(sizeof(MEFWCAPS_SKU) == 4);

//MEFWCAPS_ATTR
//    Indicates the firmware modules present in this SKU. This is an Intel
//    defined policy and it is not updateable by OEMs.
//    


/**
*
* @brief MEFWCAPS_SKU
*
*/
typedef union _MEFWCAPS_ATTR
{
   UINT32   Data;
   struct
   {
      UINT32   MeFwSize         :4;  // BITS 3:0   Size in multiples of 0.5MB
      UINT32   Reserved         :3;  // BITS 6:4
      UINT32   PbgSupport		:1;  // BIT  7     PBG Support FW
      UINT32   M3Support        :1;  // BIT  8     M3 Support
      UINT32   M0Support        :1;  // BIT  9     M0 Support
      UINT32   Reserved2        :2;  // BITS 11:10 Reserved
      UINT32   SiClass          :4;  // BITS 15:12 Si Class - All, H, M, L
      UINT32   Reserved3        :16; // BITS 31:16 Reserved
   }Fields;
}MEFWCAPS_ATTR;
C_ASSERT(sizeof(MEFWCAPS_ATTR) == 4);


#define FWCAPS_MNG_FULL_SKU_BIT    BIT0
#define FWCAPS_MNG_STD_SKU_BIT     BIT1
#define FWCAPS_AMT_SKU_BIT         BIT2
#define FWCAPS_LOCAL_MNG_SKU_BIT        BIT3
#define FWCAPS_L3_MNG_SKU_BIT         BIT4
#define FWCAPS_TDT_SKU_BIT         BIT5
#define FWCAPS_SOFTCREEK_SKU_BIT   BIT6
#define FWCAPS_VE_SKU_BIT          BIT7
#define FWCAPS_TP35_SKU_BIT        BIT8
#define FWCAPS_TP29_SKU_BIT        BIT9
#define FWCAPS_THERMREPORT_SKU_BIT BIT10
#define FWCAPS_ICC_SKU_BIT         BIT11
#define FWCAPS_PAVP_SKU_BIT        BIT12
#define FWCAPS_SPK_SKU_BIT         BIT13
#define FWCAPS_RCA_SKU_BIT         BIT14
#define FWCAPS_RPAT_SKU_BIT        BIT15
#define FWCAPS_HAP_SKU_BIT         BIT16
#define FWCAPS_IPV6_SKU_BIT        BIT17
#define FWCAPS_KVM_SKU_BIT         BIT18
#define FWCAPS_OCH_SKU_BIT         BIT19
#define FWCAPS_MEDAL_SKU_BIT        BIT20
#define FWCAPS_TLS_SKU_BIT         BIT21
#define FWCAPS_CILA_SKU_BIT        BIT22
#define FWCAPS_WLAN_SKU_BIT        BIT23
#define FWCAPS_WLDISP_SKU_BIT      BIT24
#define FWCAPS_USB3_SKU_BIT        BIT25
#define FWCAPS_NAP_SKU_BIT        BIT26
#define FWCAPS_ALARMCLK_SKU_BIT        BIT27
#define FWCAPS_MDNSPROXY_SKU_BIT     BIT30

//
#define FWCAPS_UNKNOWN_SKU_BIT     BIT31

#define FWCAPS_KERNEL_FEATURE_ID       0

#define FWCAPS_QST_FEATURE_ID          1
#define FWCAPS_ASF_FEATURE_ID          2
#define FWCAPS_AMT_FEATURE_ID          3
#define FWCAPS_AMT_FUND_FEATURE_ID     4
#define FWCAPS_TPM_FEATURE_ID          5
#define FWCAPS_DT_FEATURE_ID           6
#define FWCAPS_FPS_FEATURE_ID          7
#define FWCAPS_HOMEIT_FEATURE_ID       8
#define FWCAPS_MCTP_FEATURE_ID         9
#define FWCAPS_WOX_FEATURE_ID          10
#define FWCAPS_PMC_PATCH_FEATURE_ID    11
#define FWCAPS_VE_FEATURE_ID           12
#define FWCAPS_TDT_FEATURE_ID          13
#define FWCAPS_CORP_FEATURE_ID         14
#define FWCAPS_PLDM_FEATURE_ID         15
//
#define FWCAPS_UNKNOWN_FEATURE_ID      31

//Default settings


#define MEFWCAPS_SKU_RULE_SIZE          sizeof(MEFWCAPS_SKU)



#define MEFWCAPS_SKU_RULE_CFG           ME_RULE_CFG_CLEARED     //Not external, lockable

//
//MEFWCAPS_MANAGEABILITY_SUPP
//    Indicates the manageability support selected for the system. This is an 
//    OEM/IT/USER defined policy which can be updated and locked.
//
//    MEFWCAPS_MANAGEABILITY_SUPP_DISABLED      -  Indicates manageability is disabled.
//    MEFWCAPS_MANAGEABILITY_SUPP_AMT_ENABLED   -  Indicates AMT is enabled.
//    MEFWCAPS_MANAGEABILITY_SUPP_ASF_ENABLED   -  Indicates ASF is enabled.
//
typedef enum _MEFWCAPS_MANAGEABILITY_SUPP
{
   MEFWCAPS_MANAGEABILITY_SUPP_DISABLED = 0, // Default
   MEFWCAPS_MANAGEABILITY_SUPP_AMT_ENABLED,
   // The following two enums should be deleted after AMT_APP removes the usage from their code
   MEFWCAPS_MANAGEABILITY_SUPP_ASF_ENABLED,
   MEFWCAPS_MANAGEABILITY_SUPP_CP_ENABLED   // HomeIT
}MEFWCAPS_MANAGEABILITY_SUPP;

//Default settings
#define MEFWCAPS_MANAGEABILITY_SUPP_RULE_SIZE          4
#define MEFWCAPS_MANAGEABILITY_SUPP_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_QST_STATE
//    Indicates whether the QST must be enabled or disabled.
//
typedef enum _MEFWCAPS_QST_STATE
{
   MEFWCAPS_QST_DISABLED = 0,   
   MEFWCAPS_QST_ENABLED          // Default
}MEFWCAPS_QST_STATE;

//Default settings
#define MEFWCAPS_QST_STATE_RULE_SIZE          4
#define MEFWCAPS_QST_STATE_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_CB_STATE
//    Indicates whether the circuit breaker must be enabled or disabled.
//
typedef enum _MEFWCAPS_CB_STATE
{
   MEFWCAPS_CB_DISABLED = 0,     // Default
   MEFWCAPS_CB_ENABLED
}MEFWCAPS_CB_STATE;

//Default settings
#define MEFWCAPS_CB_STATE_RULE_SIZE          4
#define MEFWCAPS_CB_STATE_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_LAN_STATE
//    Indicates whether the LAN must be enabled or disabled.
//
typedef enum _MEFWCAPS_LAN_STATE
{
   MEFWCAPS_LAN_DISABLED = 0,    
   MEFWCAPS_LAN_ENABLED          // Default
}MEFWCAPS_LAN_STATE;

//Default settings
#define MEFWCAPS_LAN_STATE_RULE_SIZE          4
#define MEFWCAPS_LAN_STATE_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_LAN_SKU
//    Indicates the type of LAN HW SKU in use. This policy is set after a first
//    good boot and initialization of LAN driver.
//
typedef union _MEFWCAPS_LAN_SKU
{
   UINT32   Data;
   struct
   {
      UINT32   Enabled     : 1;     // Default = DISABLED
      UINT32   AsfCapable  : 1;
      UINT32   AmtCapable  : 1;
      UINT32   Reserved    : 29;
   }Fields;
}MEFWCAPS_LAN_SKU;
C_ASSERT(sizeof(MEFWCAPS_LAN_SKU) == 4);

//Default settings
#define MEFWCAPS_LAN_SKU_RULE_SIZE          sizeof(MEFWCAPS_LAN_SKU)
#define MEFWCAPS_LAN_SKU_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_ME_PLATFORM_STATE
//    Indicates whether the ME must be enabled or disabled.
//
typedef enum _MMEFWCAPS_ME_PLATFORM_STATE
{
   MEFWCAPS_ME_PLATFORM_DISABLED = 0,
   MEFWCAPS_ME_PLATFORM_ENABLED,             // Default
   MEFWCAPS_ME_PLATFORM_PASSWORD_PROTECTED   // Default
}MEFWCAPS_ME_PLATFORM_STATE;

//Default settings
#define MEFWCAPS_ME_PLATFORM_STATE_RULE_SIZE          4
#define MEFWCAPS_ME_PLATFORM_STATE_RULE_CFG           FWCAPS_RULE_STATE_CLEARED

//
//MEFWCAPS_ME_LOCAL_FW_UPDATE
//    Indicates whether the ME Local Firmware Update must be enabled or disabled.
//
typedef enum _MEFWCAPS_ME_LOCAL_FW_UPDATE
{
   MEFWCAPS_ME_LOCAL_FW_UPDATE_DISABLED = 0, // Default
   MEFWCAPS_ME_LOCAL_FW_UPDATE_ENABLED ,
   MEFWCAPS_ME_LOCAL_FW_UPDATE_PASSWORD_PROTECTED
}MEFWCAPS_ME_LOCAL_FW_UPDATE;

//Default settings
#define MEFWCAPS_ME_LOCAL_FW_UPDATE_RULE_SIZE          4
#define MEFWCAPS_ME_LOCAL_FW_UPDATE_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_TLS_STATE
//    Indicates whether TLS confidentiality is enabled or disabled.
//
#define MEFWCAPS_TLS_CONF_FOV_ENABLED  0xFF

typedef enum _MEFWCAPS_TLS_CONF_STATE
{
   MEFWCAPS_TLS_CONF_DISABLED = 0,
   MEFWCAPS_TLS_CONF_ENABLED       // Default
}MEFWCAPS_TLS_CONF_STATE;

//Default settings
#define MEFWCAPS_TLS_CONF_STATE_RULE_SIZE          4
#define MEFWCAPS_TLS_CONF_STATE_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_LCL_FW_UPD_OVR_QUAL
//    Indicates the possible values for the Local FW-Update Override Qualifier.
//
typedef enum _MEFWCAPS_LCL_FW_UPD_OVR_QUAL
{
   MEFWCAPS_FW_UPD_OVER_QUAL_ALWAYS = 0,
   MEFWCAPS_FW_UPD_OVER_QUAL_NEVER,
   MEFWCAPS_FW_UPD_OVER_QUAL_RESTRICTED
}MEFWCAPS_LCL_FW_UPD_OVR_QUAL;

//Default settings
#define MEFWCAPS_LOCAL_FW_UPD_OVR_QUAL_RULE_SIZE          4
#define MEFWCAPS_LOCAL_FW_UPD_OVR_QUAL_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_LCL_FW_UPD_OVR_CNTR
//    value defining the Local Fw-Update override counter.
//
typedef UINT32 MEFWCAPS_LCL_FW_UPD_OVR_CNTR;

//Default settings
#define MEFWCAPS_LOCAL_FW_UPD_OVR_COUNTR_RULE_SIZE         4
#define MEFWCAPS_LOCAL_FW_UPD_OVR_COUNTR_RULE_CFG          FWCAPS_RULE_LOCKABLE

// The update counter value can lie anywhere between 0 and 255
#define MEFWCAPS_LOCAL_FW_UPD_OVR_COUNTR_MAX_VALUE         0xFF 

//MEFWCAPS_OEM_SKU
//    Indicates the firmware modules present in this SKU. This is an OEM
//    defined policy 
//    

//Default settings

#define MEFWCAPS_OEM_SKU_RULE_SIZE      sizeof(MEFWCAPS_SKU)

#define MEFWCAPS_OEM_SKU_RULE_CFG       FWCAPS_RULE_STATE_CLEARED //Not external, lockable

//MEFWCAPS_LAN_BLOCK_TRAFFIC_STATE
//    Indicates whether the LAN is blocked or unblocked.
//
typedef enum _MEFWCAPS_LAN_BLOCK_TRAFFIC_STATE
{
   MEFWCAPS_LAN_TRAFFIC_UNBLOCKED = 0, // Default
   MEFWCAPS_LAN_TRAFFIC_BLOCKED
}MEFWCAPS_LAN_BLOCK_TRAFFIC_STATE;

//Default settings
#define MEFWCAPS_LAN_BLOCK_TRAFFIC_RULE_SIZE    4
#define MEFWCAPS_LAN_BLOCK_TRAFFIC_RULE_CFG     FWCAPS_RULE_STATE_CLEARED

//
//MEFWCAPS_DT_STATE
//    Indicates whether the DT must be enabled or disabled.
//
typedef enum _MEFWCAPS_DT_STATE
{
   MEFWCAPS_DT_DISABLED = 0,     // Default
   MEFWCAPS_DT_ENABLED          
}MEFWCAPS_DT_STATE;

//-----------------------------------------------------
// Platform Configuration Variables Data, added in PCH 09
//------------------------------------------------------
typedef enum _MEFWCAPS_LAN_WELL_CONFIG
{
   MEFWCAPS_LAN_CORE_WELL = 0,
   MEFWCAPS_LAN_SUS_WELL,
   MEFWCAPS_LAN_ME_WELL,
   MEFWCAPS_LAN_SLP_LAN
}MEFWCAPS_LAN_WELL_CONFIG ;

// Valid WLAN power well values are 0x80,0x82-0x85 
// MEFWCAPS_WLAN_CORE_WELL = 0x81 is reserved
typedef enum _MEFWCAPS_WLAN_WELL_CONFIG
{
   MEFWCAPS_NO_WLAN_WELL  = 0x80,
   MEFWCAPS_WLAN_SUS_WELL = 0x82,
   MEFWCAPS_WLAN_ME_WELL,
   MEFWCAPS_SLP_M_OR_SPDA, // SUS PWR DOWN ACK
   MEFWCAPS_SLP_M_OR_SMCD  //  SLP_ME_CSW_DEV
}MEFWCAPS_WLAN_WELL_CONFIG ; 

#ifndef WLAN_PWRWELL_EN
#define  MEFWCAPS_WLAN_ENABLED     MEFWCAPS_NO_WLAN_WELL
#else 
#define  MEFWCAPS_WLAN_ENABLED     MEFWCAPS_SLP_M_OR_SMCD
#endif


#if DESKTOP 
#define  MEFWCAPS_PLAT_TYPE      0x1452
#elif MOBILE 
#define  MEFWCAPS_PLAT_TYPE      0x1451
#endif

#if DEFAULT
#define  ME_FWCAPS_CPU_TYPE      0
#elif VPRO
#define  ME_FWCAPS_CPU_TYPE      1
#elif CORE
#define  ME_FWCAPS_CPU_TYPE      2
#elif CELERON
#define  ME_FWCAPS_CPU_TYPE      3
#elif UNKNOWN
#define  ME_FWCAPS_CPU_TYPE      4
#endif


#define  MEFWCAPS_CLINK_DISABLE             0
#define  MEFWCAPS_CLINK_ENABLE              1 
#define  MEFWCAPS_CLINK_OVERRIDE_DISABLE    0
#define  MEFWCAPS_CLINK_OVERRIDE_ENABLE     1

#define MEFWCAPS_MOFFOVERRIDE_DISABLE       0
#define MEFWCAPS_MOFFOVERRIDE_ENABLE        1

// c-link global disable  set:   C-link cannot be enabled
// c-link global clear: C-link can be enabled via any available option (default)
#define  MEFWCAPS_CLINK_GLOBAL_DISBALE_SET       0x1  
#define  MEFWCAPS_CLINK_GLOBAL_DISBALE_CLEAR     0x0 


typedef enum _MEFWCAPS_CPU_MISSING_LOGIC
{
   MEFWCAPS_NO_ONBOARD_GLUE_LOGIC = 0xFF
}MEFWCAPS_CPU_MISSING_LOGIC ;

typedef enum _MEFWCAPS_M3_POWER_RAILS_PRESENT
{
   MEFWCAPS_M3_PWR_RAILS_UNAVAILABLE=0,
   MEFWCAPS_M3_PWR_RAILS_AVAILABLE
}MEFWCAPS_M3_POWER_RAILS_PRESENT ;

typedef enum _MEFWCAPS_ICC_OEM_LAYOUT
{
   MEFWCAPS_BUF_THROUGH_MODE_OR_NO_MULT_SEL = 0,
}MEFWCAPS_ICC_OEM_LAYOUT ;

#define MEFWCAPS_NO_GPIO_ASSIGNED     0XFF
#define MEFWCAPS_MGPIO_PIN_ZERO           0
#define MEFWCAPS_MGPIO_PIN_ONE            1 
#define MEFWCAPS_MGPIO_PIN_TEN           10

typedef enum _MEFWCAPS_ME_EC_SPEC_COMPLIANT
{
   MEFWCAPS_NO_MEEC_IMPLEMENTATION_PRESENT =0,
   MEFWCAPS_MEEC_IMPLEMENTATION_PRESENT
}MEFWCAPS_ME_EC_SPEC_COMPLIANT ;


typedef enum _MEFWCAPS_SUS_WELL_DOWN_S45_MOFF_DC
{
 MEFWCAPS_SUS_WELL_DOWN = 0,
 MEFWCAPS_EC_CUT_SUS_WELL
}MEFWCAPS_SUS_WELL_DOWN_S45_MOFF_DC ;


//Default settings
#define MEFWCAPS_DT_RULE_SIZE          4
#define MEFWCAPS_DT_RULE_CFG           FWCAPS_RULE_LOCKABLE

//
//MEFWCAPS_CLS_STATE
//    Indicates whether CLS permit has been installed or not installed
typedef enum _MEFWCAPS_CLS_STATE
{
   MEFWCAPS_CLS_PERMIT_NOT_INSTALLED = 0,   // Default
   MEFWCAPS_CLS_PERMIT_INSTALLED
}MEFWCAPS_CLS_STATE;

typedef enum _MEFWCAPS_FOV_MANUF_STATUS
{
   MEFWCAPS_MANUF_STATUS_NOT_COMPLETE = 0,
   MEFWCAPS_MANUF_STATUS_COMPLETE,
   MEFWCAPS_MANUF_STATUS_PROCESSED  
}MEFWCAPS_FOV_MANUF_STATUS;

//Default settings

#define MEFWCAPS_OEM_CAPS_CHECK_DATA    FWCAPS_SOFTCREEK_SKU_BIT

#define MEFWCAPS_CHECK_USER_CAPS_DATA  0x005E4867

/*(FWCAPS_MNG_FULL_SKU_BIT | FWCAPS_MNG_STD_SKU_BIT | FWCAPS_AMT_SKU_BIT | \
 FWCAPS_IRWT_SKU_BIT | FWCAPS_QST_SKU_BIT | FWCAPS_TDT_SKU_BIT | FWCAPS_SOFTCREEK_SKU_BIT |  \
 FWCAPS_VE_SKU_BIT | FWCAPS_DT_SKU_BIT | FWCAPS_NAND_SKU_BIT | FWCAPS_MPC_SKU_BIT | FWCAPS_ICC_SKU_BIT | \
 FWCAPS_PAVP_SKU_BIT | FWCAPS_SPK_SKU_BIT | FWCAPS_RCA_SKU_BIT | FWCAPS_RPAT_SKU_BIT | FWCAPS_RPATCON_SKU_BIT | \
 FWCAPS_IPV6_SKU_BIT | FWCAPS_KVM_SKU_BIT | FWCAPS_OCH_SKU_BIT | FWCAPS_VLAN_SKU_BIT | \
 FWCAPS_TLS_SKU_BIT | FWCAPS_CILA_SKU_BIT)*/ 

//#define OEM_FOV_MASK  ( FWCAPS_PAVP_SKU_BIT | /* FWCAPS_NAND_SKU_BIT | */ FWCAPS_TLS_SKU_BIT | \
//                        FWCAPS_QST_SKU_BIT  | FWCAPS_SPK_SKU_BIT  | FWCAPS_IRWT_SKU_BIT | \
//                        FWCAPS_KVM_SKU_BIT  | FWCAPS_MNG_FULL_SKU_BIT | /* FWCAPS_MNG_STD_SKU_BIT |*/ \
//                        FWCAPS_AMT_SKU_BIT )

//#define FEATURE_STATE_FOV_MASK  ( FWCAPS_AMT_SKU_BIT  | FWCAPS_IRWT_SKU_BIT | \
//                                  FWCAPS_QST_SKU_BIT  | FWCAPS_PAVP_SKU_BIT | \
//                                  FWCAPS_SPK_SKU_BIT )

typedef enum _MEFWCAPS_QMQS_TO_HM_FOV_VAL
{
   MEFWCAPS_QMQS_TO_HM_NO_OVERRIDE = 0,
   MEFWCAPS_QMQS_TO_HM_OVERRIDE,
   MEFWCAPS_QMQS_TO_HM_INVALID_VAL,
} MEFWCAPS_QMQS_TO_HM_FOV_VAL;

typedef enum{

 SNB_CPU_FAMILY = 1,
 IVB_CPU_FAMILY,
 UNKNOWN_CPU_FAMILY = 0xF
}CPU_FAMILY;

typedef union _ME_PLATFORM_TYPE
{
   UINT32    Data;
   struct
   {
      UINT32   Mobile:   1;
      UINT32   Desktop:  1; 
      UINT32   Server:   1;
      UINT32   WorkStn:  1;
      UINT32   Corporate:1;
      UINT32   Consumer: 1;
      UINT32   SuperSKU: 1;
      UINT32   Rsvd:     1;
      UINT32   ImageType:4;
      UINT32   Brand:    4;
      UINT32   CpuType: 4;
      UINT32   Chipset: 4;
      UINT32   CpuBrandClass:    4;
      UINT32   PchNetInfraFuses :3;
      UINT32   Rsvd1:  1;
   }Fields;
}ME_PLATFORM_TYPE;

// Brand values
#define ME_PLATFORM_TYPE_BRAND_AMT_PRO                 1
#define ME_PLATFORM_TYPE_BRAND_STANDARD_MANAGEABILITY  2
#define ME_PLATFORM_TYPE_BRAND_L3_MANAGEABILITY        3
#define ME_PLATFORM_TYPE_BRAND_RPAT                    4
#define ME_PLATFORM_TYPE_BRAND_LOCAL_MANAGEABILITY     5
#define ME_PLATFORM_TYPE_BRAND_NO_BRAND                0

// ImageType values  
#define IMAGE_TYPE_NO_ME        0   // No ME FW
#define IMAGE_TYPE_IGNITION_FW  1   // Ignition FW 
#define IMAGE_TYPE_ME_LITE      2   // Ignition FW 
#define IMAGE_TYPE_ME_FULL_4MB  3   // ME FW 4MB image 
#define IMAGE_TYPE_ME_FULL_8MB  4   // ME FW 8MB image 



typedef enum {
    LOCAL_MNG = 1, // 0 0 1    Local Mng
    Reserved1 = 2, //0 1 0     Reserved
    Reserved2 = 3,  // 0 1 1   Reserved
    FULL_MNG = 4, // 1 0 0    Full Manageability
    STD_MNG = 5, // 1 0 1   Std Manageability
    L3_UPGRADE =  6, // 1 1 0   L3 upgrade
    NO_MNG =  7 // 1 1 1   NO manageability
}NetInfraFuses;
#pragma pack()

#endif //_FW_CAPS_MSGS_H


