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

// Disable missing documentation warnings
#pragma warning disable 1591

using System.Runtime.InteropServices;
using System;


namespace Intel.Dal
{
    [StructLayout(LayoutKind.Sequential)]
    internal struct JHI_I_EVENT_DATA
    {
        public UInt32 datalen;
        public IntPtr data;
        public JHI_EVENT_DATA_TYPE dataType;
    }
    
        [StructLayout(LayoutKind.Sequential)]
        public struct DAL_FW_VERSION
        {
            public UInt16 major;
            public UInt16 minor;
            public UInt16 hotfix;
            public UInt16 build;
        }


        [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Ansi)]
        public struct DAL_TEE_METADATA
        {
            private const int DAL_MAX_PLATFORM_TYPE_LEN = 8;
            private const int DAL_PRODUCTION_KEY_HASH_LEN = 32;
            private const int DAL_MAX_VM_TYPE_LEN = 16;
            private const int DAL_MAX_VM_VERSION_LEN = 12;
            private const int DAL_RESERVED_DWORDS = 16;

            public UInt32 api_level; // the API level of the DAL Java Class Library, unsigned integer
            public UInt32 library_version; // the version of the DAL Java Class Library for this platform, unsigned integer
            [MarshalAs(UnmanagedType.ByValTStr, SizeConst = DAL_MAX_PLATFORM_TYPE_LEN)]
            public string platform_type; // the underlying security engine on the platform, char string
            [MarshalAs(UnmanagedType.ByValArray, SizeConst = DAL_PRODUCTION_KEY_HASH_LEN)]
            public byte[] dal_key_hash; // SHA256 hash of the DAL Sign Once public key embedded in the firmware, byte array
            public UInt32 feature_set; // a bitmask of the features the platform support (SSL, NFC and etc),
            // unsigned integer bitmask vlaues in dal_feature_set_values
            [MarshalAs(UnmanagedType.ByValTStr, SizeConst = DAL_MAX_VM_TYPE_LEN)]
            public string vm_type; // the Beihai VM type in DAL, char string
            [MarshalAs(UnmanagedType.ByValTStr, SizeConst = DAL_MAX_VM_VERSION_LEN)]
            public string vm_version; // the Beihai drop version integrated into the DAL, char string
            public UInt64 access_control_groups; // a bitmask of the access control groups defined in the Java Class Library on this platform,
            // unsigned integer bitmask values in dal_access_control_groups
            public DAL_FW_VERSION fw_version; // the version of the firmware image on this platform
            [MarshalAs(UnmanagedType.ByValArray, SizeConst = DAL_RESERVED_DWORDS)]
            public UInt32[] reserved; // reserved DWORDS for future use
        }

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    internal delegate void JHI_I_CallbackFunc(IntPtr SessionHandle, [MarshalAs(UnmanagedType.Struct)] JHI_I_EVENT_DATA event_data);

    internal static class JhiWrapper
    {
        public static string INTEL_SD_UUID = "BD2FBA36A2D64DAB9390FF6DA2FEF31C";

        [StructLayout(LayoutKind.Sequential)]
        internal struct DATA_BUFFER
        {
            public IntPtr buffer;
            public UInt32 length;
            
        }

        [StructLayout(LayoutKind.Sequential)]
        public struct UUID_LIST
        {
            public UInt32 count;
            public IntPtr buffer;

        }

        [StructLayout(LayoutKind.Sequential)]
        public struct JVM_COMM_BUFFER
        {
            public DATA_BUFFER TxBuf;
            public DATA_BUFFER RxBuf;
        }

        internal const int VERSION_BUFFER_SIZE = 50;

        [StructLayout(LayoutKind.Sequential, CharSet=CharSet.Ansi)]
        internal struct JHI_I_VERSION_INFO
        {
            [MarshalAs(UnmanagedType.ByValTStr, SizeConst = VERSION_BUFFER_SIZE)]
            public string jhi_version;
            [MarshalAs(UnmanagedType.ByValTStr, SizeConst = VERSION_BUFFER_SIZE)]
            public string fw_version;
            public JHI_COMMUNICATION_TYPE comm_type;
            public JHI_PLATFROM_ID platform_id;
            [MarshalAs (UnmanagedType.ByValArray, SizeConst=20)]
            public int [] reserved;
        }

        [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Ansi)]
        internal struct JHI_I_SESSION_INFO
        {
            public JHI_SESSION_STATE state;
            public UInt32 flags;
            [MarshalAs (UnmanagedType.ByValArray, SizeConst=20)]
            public int[] reserved;
        }

        #region JHI DLL Signature checks

        [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
        internal class WINTRUST_FILE_INFO
        {
            public UInt32 cbStruct;
            public IntPtr pcwszFilePath;
            public IntPtr hFile;
            public IntPtr pgKnownSubject;
        }

        [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
        internal class WINTRUST_DATA
        {
            public UInt32 cbStruct;
            public IntPtr pPolicyCallbackData;
            public IntPtr pSIPClientData;
            public UInt32 dwUIChoice;
            public UInt32 fdwRevocationChecks;
            public UInt32 dwUnionChoice;
            public IntPtr pFile;
            public UInt32 dwStateAction;
            public IntPtr hWVTStateData;
            public IntPtr pwszURLReference;
            public UInt32 dwProvFlags;
            public UInt32 dwUIContext;
        }

        [DllImport("wintrust.dll", ExactSpelling = true, SetLastError = false, CharSet = CharSet.Unicode)]
        internal static extern uint WinVerifyTrust([In] IntPtr hwnd, [In] [MarshalAs(UnmanagedType.LPStruct)] Guid pgActionID, [In] WINTRUST_DATA pWVTData);

        [DllImport("kernel32", SetLastError = true, CharSet = CharSet.Unicode)]
        internal static extern IntPtr LoadLibrary(string lpFileName);

        // registry read functions

        [DllImport("advapi32.dll", CharSet = CharSet.Auto)]
        internal static extern int RegOpenKeyEx(UIntPtr hKey, string subKey, int options, int samDesired, out UIntPtr handle);

        [DllImport("advapi32.dll", SetLastError = true)]
        internal static extern uint RegQueryValueEx(UIntPtr hKey, string valueName, int reserved, ref uint type, System.Text.StringBuilder data, ref int dataSize);

        [DllImport("advapi32.dll", SetLastError = true)]
        internal static extern int RegCloseKey(UIntPtr hKey);

        #endregion


        private static bool is64BitProcess = (IntPtr.Size == 8); // in 64 bit process address size is 8 byte

        public static uint JHI_Initialize(out IntPtr handle, IntPtr context, UInt32 flags)
        {
            if (is64BitProcess)
                return JHI_Initialize64(out handle, context, flags);
            else
                return JHI_Initialize32(out handle, context, flags);
        }
        
        [DllImport("jhi", EntryPoint = "JHI_Initialize", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Initialize32(out IntPtr handle, IntPtr context, UInt32 flags);

        [DllImport("jhi64", EntryPoint = "JHI_Initialize", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Initialize64(out IntPtr handle, IntPtr context, UInt32 flags);

        public static uint JHI_Deinit(IntPtr handle)
        {
            if (is64BitProcess)
                return JHI_Deinit64(handle);
            else
                return JHI_Deinit32(handle);
        }

        [DllImport("jhi", EntryPoint = "JHI_Deinit", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Deinit32(IntPtr handle);

        [DllImport("jhi64", EntryPoint = "JHI_Deinit", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Deinit64(IntPtr handle);

        public static uint JHI_Uninstall(IntPtr handle, string AppId)
        {
            if (is64BitProcess)
                return JHI_Uninstall64(handle, AppId);
            else
                return JHI_Uninstall32(handle, AppId);
        }

        

        [DllImport("jhi", EntryPoint = "JHI_Uninstall", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Uninstall32(IntPtr handle, string AppId);

        [DllImport("jhi64", EntryPoint = "JHI_Uninstall", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Uninstall64(IntPtr handle, string AppId);

        public static uint JHI_GetAppletProperty(IntPtr handle, string AppId, ref JVM_COMM_BUFFER pComm)
        {
            if (is64BitProcess)
                return JHI_GetAppletProperty64(handle, AppId, ref pComm);
            else
                return JHI_GetAppletProperty32(handle, AppId, ref pComm);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetAppletProperty", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetAppletProperty32(IntPtr handle, string AppId, ref JVM_COMM_BUFFER pComm);

        [DllImport("jhi64", EntryPoint = "JHI_GetAppletProperty", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetAppletProperty64(IntPtr handle, string AppId, ref JVM_COMM_BUFFER pComm);

        // JHI GEN2 functions:

        public static uint JHI_Install2(IntPtr handle, string AppId, [In, MarshalAs(UnmanagedType.LPWStr)] string srcFile)
        {
            if (is64BitProcess)
                return JHI_Install2_64(handle, AppId, srcFile);
            else
                return JHI_Install2_32(handle, AppId, srcFile);
        }

        [DllImport("jhi", EntryPoint = "JHI_Install2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install2_32(IntPtr handle, string AppId, [In, MarshalAs(UnmanagedType.LPWStr)] string srcFile);

        [DllImport("jhi64", EntryPoint = "JHI_Install2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install2_64(IntPtr handle, string AppId, [In, MarshalAs(UnmanagedType.LPWStr)] string srcFile);

        public static uint JHI_SendAndRecv2(IntPtr handle, IntPtr SessionHandle, Int32 nCommandId, ref JVM_COMM_BUFFER pComm, ref Int32 pResponseCode)
        {
            if (is64BitProcess)
                return JHI_SendAndRecv2_64(handle, SessionHandle, nCommandId, ref pComm, ref pResponseCode);
            else
                return JHI_SendAndRecv2_32(handle, SessionHandle, nCommandId, ref pComm, ref pResponseCode);
        }

        [DllImport("jhi", EntryPoint = "JHI_SendAndRecv2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv2_32(IntPtr handle, IntPtr SessionHandle, Int32 nCommandId, ref JVM_COMM_BUFFER pComm, ref Int32 pResponseCode);

        [DllImport("jhi64", EntryPoint = "JHI_SendAndRecv2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv2_64(IntPtr handle, IntPtr SessionHandle, Int32 nCommandId, ref JVM_COMM_BUFFER pComm, ref Int32 pResponseCode);

        public static uint JHI_GetVersionInfo(IntPtr handle, ref JHI_I_VERSION_INFO VersionInfo)
        {
            if (is64BitProcess)
                return JHI_GetVersionInfo64(handle, ref VersionInfo);
            else
                return JHI_GetVersionInfo32(handle, ref VersionInfo);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetVersionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetVersionInfo32(IntPtr handle, ref JHI_I_VERSION_INFO VersionInfo);

        [DllImport("jhi64", EntryPoint = "JHI_GetVersionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetVersionInfo64(IntPtr handle, ref JHI_I_VERSION_INFO VersionInfo);

        public static uint JHI_CreateSession(IntPtr handle, string AppId, UInt32 flags, ref DATA_BUFFER initBuffer, ref IntPtr SessionHandle)
        {
            if (is64BitProcess)
                return JHI_CreateSession64(handle, AppId, flags, ref initBuffer, ref SessionHandle);
            else
                return JHI_CreateSession32(handle, AppId, flags, ref initBuffer, ref SessionHandle);
        }

        [DllImport("jhi", EntryPoint = "JHI_CreateSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_CreateSession32(IntPtr handle, string AppId, UInt32 flags, ref DATA_BUFFER initBuffer, ref IntPtr SessionHandle);

        [DllImport("jhi64", EntryPoint = "JHI_CreateSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_CreateSession64(IntPtr handle, string AppId, UInt32 flags, ref DATA_BUFFER initBuffer, ref IntPtr SessionHandle);

        public static uint JHI_CloseSession(IntPtr handle, ref IntPtr SessionHandle)
        {
            if (is64BitProcess)
                return JHI_CloseSession64(handle, ref SessionHandle);
            else
                return JHI_CloseSession32(handle, ref SessionHandle);
        }

        [DllImport("jhi", EntryPoint = "JHI_CloseSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_CloseSession32(IntPtr handle, ref IntPtr SessionHandle);

        [DllImport("jhi64", EntryPoint = "JHI_CloseSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_CloseSession64(IntPtr handle, ref IntPtr SessionHandle);

        public static uint JHI_ForceCloseSession(IntPtr handle, ref IntPtr SessionHandle)
        {
            if (is64BitProcess)
                return JHI_ForceCloseSession64(handle, ref SessionHandle);
            else
                return JHI_ForceCloseSession32(handle, ref SessionHandle);
        }

        [DllImport("jhi", EntryPoint = "JHI_ForceCloseSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_ForceCloseSession32(IntPtr handle, ref IntPtr SessionHandle);
        [DllImport("jhi64", EntryPoint = "JHI_ForceCloseSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_ForceCloseSession64(IntPtr handle, ref IntPtr SessionHandle);
       
        public static uint JHI_GetSessionsCount(IntPtr handle, string AppId, ref UInt32 SessionsCount)
        {
            if (is64BitProcess)
                return JHI_GetSessionsCount64(handle, AppId, ref SessionsCount);
            else
                return JHI_GetSessionsCount32(handle, AppId, ref SessionsCount);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetSessionsCount", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionsCount32(IntPtr handle, string AppId, ref UInt32 SessionsCount);

        [DllImport("jhi64", EntryPoint = "JHI_GetSessionsCount", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionsCount64(IntPtr handle, string AppId, ref UInt32 SessionsCount);

        public static uint JHI_GetSessionInfo(IntPtr handle, IntPtr SessionHandle, ref JHI_I_SESSION_INFO SessionInfo)
        {
            if (is64BitProcess)
                return JHI_GetSessionInfo64(handle, SessionHandle, ref SessionInfo);
            else
                return JHI_GetSessionInfo32(handle, SessionHandle, ref SessionInfo);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetSessionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionInfo32(IntPtr handle, IntPtr SessionHandle, ref JHI_I_SESSION_INFO SessionInfo);

        [DllImport("jhi64", EntryPoint = "JHI_GetSessionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionInfo64(IntPtr handle, IntPtr SessionHandle, ref JHI_I_SESSION_INFO SessionInfo);

        public static uint JHI_RegisterEvents(IntPtr handle, IntPtr SessionHandle, [In, MarshalAs(UnmanagedType.FunctionPtr)] JHI_I_CallbackFunc EventFunction)
        {
            if (is64BitProcess)
                return JHI_RegisterEvents64(handle, SessionHandle, EventFunction);
            else
                return JHI_RegisterEvents32(handle, SessionHandle, EventFunction);
        }

        [DllImport("jhi", EntryPoint = "JHI_RegisterEvents", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_RegisterEvents32(IntPtr handle, IntPtr SessionHandle, [In, MarshalAs(UnmanagedType.FunctionPtr)] JHI_I_CallbackFunc EventFunction);

        [DllImport("jhi64", EntryPoint = "JHI_RegisterEvents", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_RegisterEvents64(IntPtr handle, IntPtr SessionHandle, [In, MarshalAs(UnmanagedType.FunctionPtr)] JHI_I_CallbackFunc EventFunction);

        public static uint JHI_UnRegisterEvents(IntPtr handle, IntPtr SessionHandle)
        {
            if (is64BitProcess)
                return JHI_UnRegisterEvents64(handle, SessionHandle);
            else
                return JHI_UnRegisterEvents32(handle, SessionHandle);
        }

        [DllImport("jhi", EntryPoint = "JHI_UnRegisterEvents", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_UnRegisterEvents32(IntPtr handle, IntPtr SessionHandle);

        [DllImport("jhi64", EntryPoint = "JHI_UnRegisterEvents", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_UnRegisterEvents64(IntPtr handle, IntPtr SessionHandle);

        // JHI GEN2 functions:
        public static uint TEE_ListInstalledTAs(IntPtr handle, ref UUID_LIST appIdStrs)
        {
            return is64BitProcess
                       ? TEE_ListInstalledTAs64(handle, ref appIdStrs)
                       : TEE_ListInstalledTAs32(handle, ref appIdStrs);
        }

        [DllImport("TeeManagement", EntryPoint = "TEE_ListInstalledTAs", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_ListInstalledTAs32(IntPtr handle, [MarshalAs(UnmanagedType.Struct) ] ref UUID_LIST appIdStrs);
        [DllImport("TeeManagement64", EntryPoint = "TEE_ListInstalledTAs", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_ListInstalledTAs64(IntPtr handle, [MarshalAs(UnmanagedType.Struct)] ref UUID_LIST appIdStrs);
        
        public static uint TEE_OpenSDSession(string sdId, out IntPtr sdHandle)
        {
            return is64BitProcess ? TEE_OpenSDSession64(sdId, out sdHandle) : TEE_OpenSDSession32(sdId, out sdHandle);
        }

        [DllImport("TeeManagement", EntryPoint = "TEE_OpenSDSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_OpenSDSession32(string sdId, out IntPtr sdHandle);
        [DllImport("TeeManagement64", EntryPoint = "TEE_OpenSDSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_OpenSDSession64(string sdId, out IntPtr sdHandle);

        public static uint TEE_CloseSDSession(ref IntPtr sdHandle)
        {           
            return is64BitProcess ? TEE_CloseSDSession64(ref sdHandle) : TEE_CloseSDSession32(ref sdHandle);
        }

        [DllImport("TeeManagement", EntryPoint = "TEE_CloseSDSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_CloseSDSession32(ref IntPtr sdHandle);
        [DllImport("TeeManagement64", EntryPoint = "TEE_CloseSDSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_CloseSDSession64(ref IntPtr sdHandle);

        public static uint TEE_SendAdminCmdPkg(IntPtr handle, IntPtr package, int length)
        {
            return is64BitProcess ? TEE_SendAdminCmdPkg64(handle, package, length) : TEE_SendAdminCmdPkg32(handle, package, length);
        }

        [DllImport("TeeManagement", EntryPoint = "TEE_SendAdminCmdPkg", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_SendAdminCmdPkg32(IntPtr handle, IntPtr package, int length);
        [DllImport("TeeManagement64", EntryPoint = "TEE_SendAdminCmdPkg", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_SendAdminCmdPkg64(IntPtr handle, IntPtr package, int length);
        
        public static uint JHI_TEE_SendAdminCmd(byte[] bytes)
        {
            uint result = 0;

            IntPtr sdHandle;
            uint res = TEE_OpenSDSession(INTEL_SD_UUID, out sdHandle);

            if (res != 0)
                return res;
            
            bytes = bytes ?? new byte[0];

            int numBytes = bytes.Length;
            IntPtr pBytes = Marshal.AllocHGlobal(numBytes);
            Marshal.Copy(bytes, 0, pBytes, numBytes);
            res = TEE_SendAdminCmdPkg(sdHandle, pBytes, numBytes);
            Marshal.FreeHGlobal(pBytes);

            if (res != 0)
                result = res;

            res = TEE_CloseSDSession(ref sdHandle);

            if (res != 0)
                result = res;

            return result;
        }
        // JHI GEN2 functions:
        public static uint TEE_QueryTEEMetadata(IntPtr handle, ref DAL_TEE_METADATA dalMetaData)
        {
            return is64BitProcess
                       ? TEE_QueryTEEMetadata64(handle, ref dalMetaData)
                       : TEE_QueryTEEMetadata32(handle, ref dalMetaData);
        }

        [DllImport("TeeManagement", EntryPoint = "TEE_QueryTEEMetadata", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_QueryTEEMetadata32(IntPtr handle, [MarshalAs(UnmanagedType.Struct)] ref DAL_TEE_METADATA dalMetaData);
        [DllImport("TeeManagement64", EntryPoint = "TEE_QueryTEEMetadata", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint TEE_QueryTEEMetadata64(IntPtr handle, [MarshalAs(UnmanagedType.Struct)] ref DAL_TEE_METADATA dalMetaData);
        
    }
}

// Restore missing documentation warnings
#pragma warning restore 1591
