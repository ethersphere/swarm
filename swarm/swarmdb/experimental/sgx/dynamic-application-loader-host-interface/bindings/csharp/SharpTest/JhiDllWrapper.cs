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

﻿using System.Runtime.InteropServices;
using System;
using Intel.Dal;


namespace SharpTest
{
    [StructLayout(LayoutKind.Sequential)]
    public struct JHI_I_EVENT_DATA
    {
        public UInt32 datalen;
        public IntPtr data;
        public JHI_EVENT_DATA_TYPE dataType;
    }

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    public delegate void JHI_I_CallbackFunc(IntPtr SessionHandle, [MarshalAs(UnmanagedType.Struct)] JHI_I_EVENT_DATA event_data);

    public class JhiDllWrapper
    {
        #region structs
        [StructLayout(LayoutKind.Sequential)]
        public struct DATA_BUFFER
        {
            public IntPtr buffer;
            public UInt32 length;

        }

        [StructLayout(LayoutKind.Sequential)]
        public struct JVM_COMM_BUFFER
        {
            public DATA_BUFFER TxBuf;
            public DATA_BUFFER RxBuf;
        }

        internal const int VERSION_BUFFER_SIZE = 50;

        [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Ansi)]
        public struct JHI_I_VERSION_INFO
        {
            [MarshalAs(UnmanagedType.ByValTStr, SizeConst = VERSION_BUFFER_SIZE)]
            public string jhi_version;
            [MarshalAs(UnmanagedType.ByValTStr, SizeConst = VERSION_BUFFER_SIZE)]
            public string fw_version;
            public JHI_COMMUNICATION_TYPE comm_type;
            public JHI_PLATFROM_ID platform_id;
            [MarshalAs(UnmanagedType.ByValArray, SizeConst = 20)]
            public int[] reserved;
        }

        [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Ansi)]
        public struct JHI_I_SESSION_INFO
        {
            public JHI_SESSION_STATE state;
            public UInt32 flags;
            [MarshalAs(UnmanagedType.ByValArray, SizeConst = 20)]
            public int[] reserved;
        }

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

        public static uint JHI_SendAndRecv(IntPtr handle, string AppId, ref JVM_COMM_BUFFER pComm)
        {
            if (is64BitProcess)
                return JHI_SendAndRecv64(handle, AppId, ref pComm);
            else
                return JHI_SendAndRecv32(handle, AppId, ref pComm);
        }

        [DllImport("jhi", EntryPoint = "JHI_SendAndRecv", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv32(IntPtr handle, string AppId, ref JVM_COMM_BUFFER pComm);

        [DllImport("jhi64", EntryPoint = "JHI_SendAndRecv", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv64(IntPtr handle, string AppId, ref JVM_COMM_BUFFER pComm);

     
        public static uint JHI_Install(IntPtr handle, string AppId, [In, MarshalAs(UnmanagedType.LPWStr)] string srcFile)
        {
            if (is64BitProcess)
                return JHI_Install64(handle, AppId, srcFile);
            else
                return JHI_Install32(handle, AppId, srcFile);
        }

        [DllImport("jhi", EntryPoint = "JHI_Install", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install32(IntPtr handle, string AppId, [In, MarshalAs(UnmanagedType.LPWStr)] string srcFile);

        [DllImport("jhi64", EntryPoint = "JHI_Install", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install64(IntPtr handle, string AppId, [In, MarshalAs(UnmanagedType.LPWStr)] string srcFile);


        
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

        public static uint JHI_SendAndRecv2(IntPtr handle, IntPtr SessionHandle, UInt32 nCommandId, ref JVM_COMM_BUFFER pComm, ref UInt32 pResponseCode)
        {
            if (is64BitProcess)
                return JHI_SendAndRecv2_64(handle, SessionHandle, nCommandId, ref pComm, ref pResponseCode);
            else
                return JHI_SendAndRecv2_32(handle, SessionHandle, nCommandId, ref pComm, ref pResponseCode);
        }

        [DllImport("jhi", EntryPoint = "JHI_SendAndRecv2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv2_32(IntPtr handle, IntPtr SessionHandle, UInt32 nCommandId, ref JVM_COMM_BUFFER pComm, ref UInt32 pResponseCode);

        [DllImport("jhi64", EntryPoint = "JHI_SendAndRecv2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv2_64(IntPtr handle, IntPtr SessionHandle, UInt32 nCommandId, ref JVM_COMM_BUFFER pComm, ref UInt32 pResponseCode);

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


        #region low level methods
      
        public static uint JHI_SendAndRecv(IntPtr handle, IntPtr AppId, IntPtr pComm)
        {
            if (is64BitProcess)
                return JHI_SendAndRecv64(handle, AppId, pComm);
            else
                return JHI_SendAndRecv32(handle, AppId, pComm);
        }

        [DllImport("jhi", EntryPoint = "JHI_SendAndRecv", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv32(IntPtr handle, IntPtr AppId, IntPtr pComm);

        [DllImport("jhi64", EntryPoint = "JHI_SendAndRecv", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv64(IntPtr handle, IntPtr AppId, IntPtr pComm);


      

        public static uint JHI_Install(IntPtr handle, IntPtr AppId, IntPtr srcFile)
        {
            if (is64BitProcess)
                return JHI_Install64(handle, AppId, srcFile);
            else
                return JHI_Install32(handle, AppId, srcFile);
        }

        [DllImport("jhi", EntryPoint = "JHI_Install", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install32(IntPtr handle, IntPtr AppId, IntPtr srcFile);

        [DllImport("jhi64", EntryPoint = "JHI_Install", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install64(IntPtr handle, IntPtr AppId, IntPtr srcFile);



        public static uint JHI_Uninstall(IntPtr handle, IntPtr AppId)
        {
            if (is64BitProcess)
                return JHI_Uninstall64(handle, AppId);
            else
                return JHI_Uninstall32(handle, AppId);
        }

        [DllImport("jhi", EntryPoint = "JHI_Uninstall", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Uninstall32(IntPtr handle, IntPtr AppId);

        [DllImport("jhi64", EntryPoint = "JHI_Uninstall", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Uninstall64(IntPtr handle, IntPtr AppId);



        public static uint JHI_GetAppletProperty(IntPtr handle, IntPtr AppId, IntPtr pComm)
        {
            if (is64BitProcess)
                return JHI_GetAppletProperty64(handle, AppId, pComm);
            else
                return JHI_GetAppletProperty32(handle, AppId, pComm);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetAppletProperty", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetAppletProperty32(IntPtr handle, IntPtr AppId, IntPtr pComm);

        [DllImport("jhi64", EntryPoint = "JHI_GetAppletProperty", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetAppletProperty64(IntPtr handle, IntPtr AppId, IntPtr pComm);


        // JHI GEN2 functions:

        public static uint JHI_Install2(IntPtr handle, IntPtr AppId, IntPtr srcFile)
        {
            if (is64BitProcess)
                return JHI_Install2_64(handle, AppId, srcFile);
            else
                return JHI_Install2_32(handle, AppId, srcFile);
        }

        [DllImport("jhi", EntryPoint = "JHI_Install2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install2_32(IntPtr handle, IntPtr AppId,IntPtr  srcFile);

        [DllImport("jhi64", EntryPoint = "JHI_Install2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_Install2_64(IntPtr handle, IntPtr AppId, IntPtr  srcFile);

        public static uint JHI_SendAndRecv2(IntPtr handle, IntPtr SessionHandle, UInt32 nCommandId, IntPtr pComm, IntPtr pResponseCode)
        {
            if (is64BitProcess)
                return JHI_SendAndRecv2_64(handle, SessionHandle, nCommandId,  pComm,  pResponseCode);
            else
                return JHI_SendAndRecv2_32(handle, SessionHandle, nCommandId,  pComm,  pResponseCode);
        }

        [DllImport("jhi", EntryPoint = "JHI_SendAndRecv2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv2_32(IntPtr handle, IntPtr SessionHandle, UInt32 nCommandId, IntPtr pComm, IntPtr pResponseCode);

        [DllImport("jhi64", EntryPoint = "JHI_SendAndRecv2", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_SendAndRecv2_64(IntPtr handle, IntPtr SessionHandle, UInt32 nCommandId, IntPtr pComm, IntPtr pResponseCode);

        public static uint JHI_GetVersionInfo(IntPtr handle, IntPtr VersionInfo)
        {
            if (is64BitProcess)
                return JHI_GetVersionInfo64(handle,  VersionInfo);
            else
                return JHI_GetVersionInfo32(handle,  VersionInfo);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetVersionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetVersionInfo32(IntPtr handle, IntPtr VersionInfo);

        [DllImport("jhi64", EntryPoint = "JHI_GetVersionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetVersionInfo64(IntPtr handle, IntPtr VersionInfo);

        public static uint JHI_CreateSession(IntPtr handle, IntPtr AppId, UInt32 flags, IntPtr initBuffer, ref IntPtr SessionHandle)
        {
            if (is64BitProcess)
                return JHI_CreateSession64(handle, AppId, flags,   initBuffer, ref SessionHandle);
            else
                return JHI_CreateSession32(handle, AppId, flags,   initBuffer, ref SessionHandle);
        }

        [DllImport("jhi", EntryPoint = "JHI_CreateSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_CreateSession32(IntPtr handle, IntPtr AppId, UInt32 flags, IntPtr initBuffer, ref IntPtr SessionHandle);

        [DllImport("jhi64", EntryPoint = "JHI_CreateSession", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_CreateSession64(IntPtr handle, IntPtr AppId, UInt32 flags, IntPtr initBuffer, ref IntPtr SessionHandle);

        
        public static uint JHI_GetSessionsCount(IntPtr handle, IntPtr AppId, IntPtr SessionsCount)
        {
            if (is64BitProcess)
                return JHI_GetSessionsCount64(handle, AppId,  SessionsCount);
            else
                return JHI_GetSessionsCount32(handle, AppId,  SessionsCount);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetSessionsCount", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionsCount32(IntPtr handle, IntPtr AppId, IntPtr SessionsCount);

        [DllImport("jhi64", EntryPoint = "JHI_GetSessionsCount", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionsCount64(IntPtr handle, IntPtr AppId, IntPtr SessionsCount);

        public static uint JHI_GetSessionInfo(IntPtr handle, IntPtr SessionHandle, IntPtr SessionInfo)
        {
            if (is64BitProcess)
                return JHI_GetSessionInfo64(handle, SessionHandle,   SessionInfo);
            else
                return JHI_GetSessionInfo32(handle, SessionHandle,   SessionInfo);
        }

        [DllImport("jhi", EntryPoint = "JHI_GetSessionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionInfo32(IntPtr handle, IntPtr SessionHandle, IntPtr SessionInfo);

        [DllImport("jhi64", EntryPoint = "JHI_GetSessionInfo", CallingConvention = CallingConvention.Cdecl)]
        private static extern uint JHI_GetSessionInfo64(IntPtr handle, IntPtr SessionHandle, IntPtr SessionInfo);
        
        #endregion

        #region Helper Methods

        /// <summary>
        /// Removes all the '-' characters from an applet id.
        /// </summary>
        /// <param name="AppId">Applet ID</param>
        /// <returns>Applet ID without '-' chars</returns>
        public static string RemoveSeperators(string AppId)
        {
            return AppId.Replace("-", "");
        }



        #endregion
    }
}
