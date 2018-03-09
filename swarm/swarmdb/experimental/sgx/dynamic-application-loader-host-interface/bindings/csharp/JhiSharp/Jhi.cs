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

using System;
using System.Runtime.InteropServices;
using System.Diagnostics;
using System.Collections.Generic;
using Microsoft.Win32;

namespace Intel.Dal
{
    /// <summary>
    /// this enum lists the flags that used when creating a session
    /// </summary>
    [Flags]
    public enum JHI_SESSION_FLAGS
    {
        /// <summary>
        /// no flags to be used
        /// </summary>
        None = 0,

        /// <summary>
        /// create a shared session, or receive a handle for an existing shared session
        /// </summary>
        SharedSession = 1
    }

    /// <summary>
    /// this enum lists the communication types that are used
    /// by JHI in order to communicate with the firmware
    /// </summary>
    public enum JHI_COMMUNICATION_TYPE
    {
        /// <summary>
        /// communication by sockets
        /// </summary>
        JHI_SOCKETS = 0,

        /// <summary>
        /// communication by HECI
        /// </summary>
        JHI_HECI = 1
    }

    /// <summary>
    /// this enum lists the platfom types that are supported by JHI
    /// </summary>
    public enum JHI_PLATFROM_ID
    {
        /// <summary>
        /// Intel(R) Management Engine (Intel(R) ME)
        /// </summary>
        ME = 0,
        /// <summary>
        /// VLV
        /// </summary>
        SEC = 1,
        /// <summary>
        /// CSE
        /// </summary>
        CSE = 2,
        /// <summary>
        /// invalid platform
        /// </summary>
        INVALID_PLATFORM_ID = 3
    }

    /// <summary>
    /// this struct contains global information of JHI such as JHI version and the FW version
    /// which can be used by applocations to determine DAL capabilities 
    /// </summary>
    public struct JHI_VERSION_INFO
    {
        /// <summary>
        /// the version of the JHI service in format: Major.Minor.Hotfix.Build 
        /// </summary>
        public string jhi_version;

        /// <summary>
        /// the version of the firmware in format: Major.Minor.Hotfix.Build
        /// </summary>
        public string fw_version;

        /// <summary>
        /// the communication type between JHI and the firmware
        /// </summary>
        public JHI_COMMUNICATION_TYPE comm_type;

        /// <summary>
        /// the platform supported by the JHI service
        /// </summary>
        public JHI_PLATFROM_ID platform_id;
    }

    /// <summary>
    /// this enum lists the states of a session
    /// </summary>
    public enum JHI_SESSION_STATE
    {
        /// <summary>
        /// the session is active
        /// </summary>
        JHI_SESSION_STATE_ACTIVE = 0,

        /// <summary>
        /// the session does not exists
        /// </summary>
        JHI_SESSION_STATE_NOT_EXISTS = 1
    }

    /// <summary>
    /// this struct contains information for a given session
    /// </summary>
    public struct JHI_SESSION_INFO
    {
        /// <summary>
        /// the session state
        /// </summary>
        public JHI_SESSION_STATE state;

        /// <summary>
        /// the flags used when this session created
        /// </summary>
        public JHI_SESSION_FLAGS flags;
    } 


    /// <summary>
    /// This is the main class that is used in order to communicate with Intel(R) DAL via DAL Host Interface service (JHI)
    /// </summary>
    public class Jhi
    {
        private static volatile Jhi instance;
        private static object syncLock = new object();

        private IntPtr _handle;
        private object _eventsLock = new object();

        private List<JhiSession> _eventList;

        private const uint DEFUALT_APPLET_PROPERTY_LEN = 512; // applet property lenght in characters (1K size)

        private JHI_I_CallbackFunc JhiMainCallback;

        private static bool _disableDllValidation = false;
        
        /// <summary>
        ///     This flags is used in order to disable JHI.DLL signature validation.
        ///     Warning: Production applications should not use this flag and leave it as is. Disabling JHI.DLL signature validation
        ///     will result with a security hole since the JhiSharp might load a malicious DLL.
        /// </summary>
        public static bool DisableDllValidation
        {
            get { return _disableDllValidation; }
            set { _disableDllValidation = value; }
        }

        /// <summary>
        /// Constructor
        /// </summary>
        private Jhi()
        {
            uint ret;
            _handle = new IntPtr();

            if (!DisableDllValidation)
                validateJhiDllSignature();

            ret = JhiWrapper.JHI_Initialize(out _handle, IntPtr.Zero, 0);

            if (ret != 0)
                throw new JhiException("JHI_Initialize() failed", ret);

            _eventList = new List<JhiSession>();
            JhiMainCallback = this.CallbackFunc;
        }

        private string getJhiDllPath()
        {
            string jhiPath = "";
            bool is64BitProcess = (IntPtr.Size == 8); // in 64 bit applications, process address size is 8 byte


            if (is64BitProcess)
            {
                // in 64 bit OS and application. we need to read the Program files location
                // that contains jhi64.dll
                string ProgramsDir = Environment.GetEnvironmentVariable("ProgramW6432");
                
                jhiPath = System.IO.Path.Combine(ProgramsDir, @"Intel\Intel(R) Management Engine Components\DAL\jhi64.dll");
            }
            else
            {
                // jhi installation path can vary according to the MEI installer, therefore we have to 
                // retrieve the installation folder from the registry in order to get the location of jhi.dll

                UIntPtr LOCAL_MACHINE = new UIntPtr(0x80000002u);
                int KEY_READ = 0x20019;
                int KEY_WOW64_64KEY = 0x0100;
                UIntPtr handle;
                uint varaiableType = 0; // REG_SZ
                int size = (260-1)*2; // (FILENAME_MAX-1) * sizeof(w_char)
                System.Text.StringBuilder keyBuffer = new System.Text.StringBuilder(size);

                if (JhiWrapper.RegOpenKeyEx(LOCAL_MACHINE, @"SOFTWARE\Intel\Services\DAL", 0, KEY_READ | KEY_WOW64_64KEY, out handle) == 0)
                {
                    if (JhiWrapper.RegQueryValueEx(handle, "FILELOCALE", 0, ref varaiableType, keyBuffer, ref size) == 0)
                    {
                        jhiPath = System.IO.Path.Combine(keyBuffer.ToString(),"Jhi.dll");
                    }

                    // close the registry handle
                    JhiWrapper.RegCloseKey(handle);
                }

            }

            return jhiPath;
        }
        
        /// <summary>
        ///     This function performs the following:
        ///     1) verify JHI DLL exist is in the DAL folder
        ///     2) verify JHI DLL is singed by Intel
        ///     3) load the singed DLL
        /// </summary>
        private void validateJhiDllSignature()
        {
            string jhiFileName = getJhiDllPath();

            //Verify File Exists
            if (!System.IO.File.Exists(jhiFileName))
                throw new System.IO.FileNotFoundException("Could not find JHI DLL - make sure Intel(R) Dynamic Application Loader Host Interface Service is installed", jhiFileName);


            JhiWrapper.WINTRUST_FILE_INFO fileInfo = new JhiWrapper.WINTRUST_FILE_INFO();
            JhiWrapper.WINTRUST_DATA wintrustData = new JhiWrapper.WINTRUST_DATA();

            try
            {
                //Verify File Signature

                fileInfo.cbStruct = (UInt32)Marshal.SizeOf(typeof(JhiWrapper.WINTRUST_FILE_INFO));
                fileInfo.pcwszFilePath = Marshal.StringToCoTaskMemAuto(jhiFileName);
                fileInfo.hFile = IntPtr.Zero;
                fileInfo.pgKnownSubject = IntPtr.Zero;

                wintrustData.cbStruct = (UInt32)Marshal.SizeOf(typeof(JhiWrapper.WINTRUST_DATA));
                wintrustData.pPolicyCallbackData = IntPtr.Zero;
                wintrustData.pSIPClientData = IntPtr.Zero;
                wintrustData.dwUIChoice = 2;
                wintrustData.fdwRevocationChecks = 0;
                wintrustData.dwUnionChoice = 1;
                wintrustData.dwStateAction = 0;
                wintrustData.hWVTStateData = IntPtr.Zero;
                wintrustData.pwszURLReference = IntPtr.Zero;
                wintrustData.dwUIContext = 0;

                wintrustData.pFile = Marshal.AllocCoTaskMem(Marshal.SizeOf(typeof(JhiWrapper.WINTRUST_FILE_INFO)));
                Marshal.StructureToPtr(fileInfo, wintrustData.pFile, false);

                Guid guid = new Guid("{00AAC56B-CD44-11d0-8CC2-00C04FC295EE}");
                IntPtr invalidHandle = new IntPtr(-1);

                uint ret = JhiWrapper.WinVerifyTrust(invalidHandle, guid, wintrustData);

                //Success = 0,
                //ProviderUnknown = 0x800b0001,           
                //ActionUnknown = 0x800b0002,         
                //SubjectFormUnknown = 0x800b0003,       
                //SubjectNotTrusted = 0x800b0004,     
                //FileNotSigned = 0x800B0100,         
                //SubjectExplicitlyDistrusted = 0x800B0111,   
                //SignatureOrFileCorrupt = 0x80096010,    
                //SubjectCertExpired = 0x800B0101,        
                //SubjectCertificateRevoked = 0x800B010      

                if (ret != 0)
                    throw new Exception();


                //Verify File Publisher

                // CN=Intel Corporation, OU=ISWQL, OU=Digital ID Class 3 - Microsoft Software Validation v2, O=Intel Corporation, L=Folsom, S=California, C=US

                System.Security.Cryptography.X509Certificates.X509Certificate2 cert = new System.Security.Cryptography.X509Certificates.X509Certificate2(jhiFileName);

                if (!cert.Subject.Contains("CN=Intel Corporation"))
                    throw new Exception();


                // Load the DLL manualy to make sure it is loaded from the right location.
                if (JhiWrapper.LoadLibrary(jhiFileName) == IntPtr.Zero)
                    throw new Exception();
            }
            catch
            {
                throw new System.Security.SecurityException("Could not verify DLL signature of '" + jhiFileName + "'");
            }
            finally
            {
                Marshal.FreeCoTaskMem(wintrustData.pFile);
                Marshal.FreeCoTaskMem(fileInfo.pcwszFilePath);
            }

        }

        /// <summary>
        /// Buffer size limitation in JHI requests is 2MB, 
        /// JHI will not accept any buffer with greater size.
        ///
        /// Note that this size limitiation does not mark the maximum buffer size an applet can recieve,
        /// applet max buffer size changes from one applet to another.
        ///
        /// This applies for all JHI API function that use buffers such as: 
        /// <see cref="Jhi.SendAndRecv2"/>, <see cref="Jhi.CreateSession"/>.
        /// </summary>
        public const int JHI_BUFFER_MAX = 2097152;

        /// <summary>
        /// While applet version is represented in a Major.Minor format (i.e. 1.0)
        /// the VM repersntation of an applet version (that can be obtained using JHI_GetAppletProperty) is as an integer that combine both major and minor version.
        /// in order to perform the transition between to two representation we offer the following macros:
        /// This function create a VM Applet Version (32bit) from a Major.Minor format
        /// 
		/// Bits:
        ///         00-07 - Major
        ///         08-15 - Minor
        ///         15-31 – Reserved (All Zero)
        /// </summary>
        /// <param name="maj">Applet Major Version</param>
        /// <param name="min">Applet Minor Version</param>
        /// <returns>the applet version in VM format (integer) </returns>
        public static UInt32 MakeAppletVersion(UInt32 maj,UInt32 min)
        {
            return (UInt32) ((maj & 0x000000FFUL) | ((min << 8) & 0x0000FF00UL) & (0x0000FFFFUL)); 
        }

        /// <summary>
        ///     Extract Applet Major Version from a VM integer representation (num)
        /// </summary>
        /// <param name="num">applet version in VM integer representation</param>
        /// <returns>Applet Major Version</returns>
        public static Byte MajorAppletVer(UInt32 num)
        {
            return (Byte) (num & 0x000000FFUL); 
        }

        /// <summary>
        ///     Extract Applet Minor Version from a VM integer representation (num)
        /// </summary>
        /// <param name="num">applet version in VM integer representation</param>
        /// <returns>Applet Minor Version</returns>
        public static Byte MinorAppletVer(UInt32 num)
        {
            return (Byte)((num & 0x0000FF00UL) >> 8); 
        }

        /// <summary>
        /// An instance of the JHI singelton class.
        /// The first call to this member will invoke JHI initialization, in case of error JhiException is thrown. 
        /// </summary>
        /// <exception cref="JhiException">
        /// Thrown when JHI initialization failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_ERROR_REGISTRY"/></term><description>JHI failed to read/write form the registry</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_ERROR_REPOSITORY_NOT_FOUND"/></term><description>The applets repository directory wasn’t found</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SPOOLER_NOT_FOUND"/></term><description>The spooler applet file wasn’t found</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_SPOOLER"/></term><description>Cannot download spooler / create an instance of the spooler</description></item>
        /// </list>
        /// </exception>
        public static Jhi Instance
        {
              get 
              {
                 if (instance == null) 
                 {
                    lock (syncLock) 
                    {
                       if (instance == null)
                           instance = new Jhi();
                    }
                 }

                 return instance;
              }
        }
        
        /// <summary>
        ///     This finalizer performs deinit for JHI, releasing
        ///     resources allocated by initialization.
        /// </summary>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// </list>
        /// </exception>
        ~Jhi()
        {
            if (_handle != null)
            {
                uint ret;
                ret = JhiWrapper.JHI_Deinit(_handle);
                Debug.Assert(ret == 0);
            }
        }

        #region API functions

        /// <summary>
        /// Install Applet into the Intel(R) DAL
        /// </summary>
        /// <param name="AppId">Applet ID</param>
        /// <param name="srcFile">Applet filename Path</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_FILE_ERROR_COPY"/></term><description>Failed to copy the DALP file to the repository</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_FILE_NOT_FOUND"/></term><description>The DALP file does not exist</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_READ_FROM_FILE_FAILED"/></term><description>Failed to open the DALP file for read</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_PACKAGE_FORMAT"/></term><description>The DALP file has invalid format</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INSTALL_FAILED"/></term><description>no compatible applet was found in the DALP file</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_MAX_INSTALLED_APPLETS_REACHED"/></term><description>exceeded max applets allowed</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INSTALL_FAILURE_SESSIONS_EXISTS"/></term><description>cannot install while there are active sessions</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_FILE_EXTENSION"/></term><description>Applet files must end with .dalp extension</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_INSTALL_FILE"/></term><description>The applet file path is invalid</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_APPLET_GUID"/></term><description>The applet id is invalid</description></item>
        /// </list>
        /// </exception>
        public void Install(string AppId, string srcFile)
        {
            uint ret;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (AppId == null)
            {
                throw new ArgumentNullException("AppId");
            }

            if (srcFile == null)
            {
                throw new ArgumentNullException("path");
            }

            ret = JhiWrapper.JHI_Install2(_handle, RemoveSeperators(AppId), srcFile);
            if (ret != 0)
                throw new JhiException("JHI_Install2() failed", ret);
            
        }

        /// <summary>
        /// Uninstall applet from the Intel(R) DAL
        /// </summary>
        /// <param name="AppId">Applet ID</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_APPLET_GUID"/></term><description>The applet id is invalid</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_UNINSTALL_FAILURE_SESSIONS_EXISTS"/></term><description>Cannot uninstall applet while sessions exists</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_DELETE_FROM_REPOSITORY_FAILURE"/></term><description>Failed to remove the applet file form the repository</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_APPLET_NOT_INSTALLED"/></term><description>The applet is not installed</description></item>
        /// </list>
        /// </exception>
        public void Uninstall(string AppId)
        {
            uint ret;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (AppId == null)
            {
                throw new ArgumentNullException("AppId");
            }

            ret = JhiWrapper.JHI_Uninstall(_handle, RemoveSeperators(AppId));
            if (ret != 0)
                throw new JhiException("JHI_Uninstall() failed", ret);

        }

        /// <summary>
        /// send and receive data from the application to an applet session
        /// </summary>
        /// <param name="Session">Session Handle</param>
        /// <param name="nCommandId">Command ID to send the data to</param>
        /// <param name="InBuf">Input Data</param>
        /// <param name="OutBuf">Output Data</param>
        /// <param name="ResponseCode">an error code that is returned from the applet</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to the Session argument</exception>
        /// <exception cref="JhiInsufficientBufferException">Thrown when OutBuf is too short to contain the response. The required size reside within the class Required_size member</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_SESSION_HANDLE"/></term><description>The session handle is not valid</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_APPLET_FATAL"/></term><description>Returned when the applet session has crashed</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_BUFFER_SIZE"/></term><description>Used a buffer that is larger than <see cref="Jhi.JHI_BUFFER_MAX"/></description></item>
        /// </list>
        /// </exception>
        public void SendAndRecv2(JhiSession Session, Int32 nCommandId, byte[] InBuf, ref byte[] OutBuf, out Int32 ResponseCode)
        {
            uint ret;
            bool useEmptyOutBuffer = false;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (Session == null)
                throw new ArgumentNullException("Session");
            

            if (InBuf == null)
                InBuf = new byte[0];

            if (OutBuf == null) // allow null buffer for output
            {
                OutBuf = new byte[0]; // temporary convert the out buffer to a byte[0] array.
                useEmptyOutBuffer = true;
            }

            ResponseCode = 0;

            IntPtr tmpInBuf = Marshal.AllocHGlobal(InBuf.Length);

            if (InBuf.Length > 0)
                Marshal.Copy(InBuf, 0, tmpInBuf, InBuf.Length);

            IntPtr tmpOutBuf = Marshal.AllocHGlobal(OutBuf.Length);

            JhiWrapper.JVM_COMM_BUFFER comm;
            comm.TxBuf.buffer = tmpInBuf;
            comm.TxBuf.length = (uint)InBuf.Length;
            comm.RxBuf.buffer = tmpOutBuf;
            comm.RxBuf.length = (uint)OutBuf.Length;

            ret = JhiWrapper.JHI_SendAndRecv2(_handle, Session.SessionHandle, nCommandId, ref comm,ref ResponseCode);

            if (useEmptyOutBuffer)
                OutBuf = null;

            if (ret != 0)
            {
                Marshal.FreeHGlobal(tmpInBuf);
                Marshal.FreeHGlobal(tmpOutBuf);

                if (ret == 0x200) // insufficient buffer
                {
                    throw new JhiInsufficientBufferException("JHI_SendAndRecv failed: insufficient output buffer. required size:" + comm.RxBuf.length, ret, comm.RxBuf.length);
                }
                else throw new JhiException("JHI_SendAndRecv() failed", ret);
            }

            if (comm.RxBuf.length >= 0)
            {
                OutBuf = new byte[comm.RxBuf.length];

                if (comm.RxBuf.length > 0)
                    Marshal.Copy(comm.RxBuf.buffer, OutBuf, 0, (int)comm.RxBuf.length);
            }

            Marshal.FreeHGlobal(tmpInBuf);
            Marshal.FreeHGlobal(tmpOutBuf);
        }

        /// <summary>
        /// Get a property value of an installed applet
        /// </summary>
        /// <param name="AppId">The applet uuid</param>
        /// <param name="Property">the property name, wich can be one of the following strings:
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible applet properties</term>
        /// </listheader>
        /// <item><term>applet.name</term></item>
        /// <item><term>applet.vendor</term></item>
        /// <item><term>applet.description</term></item>
        /// <item><term>applet.version</term></item>
        /// <item><term>security.version</term></item>
        /// <item><term>applet.flash.quota</term></item>
        /// <item><term>applet.debug.enable</term></item>
        /// <item><term>applet.shared.session.support</term></item>
        /// <item><term>applet.platform</term></item>
        /// </list>
        /// 
        /// </param>
        /// <param name="Value">the output value</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_APPLET_NOT_INSTALLED"/></term><description>The applet is not installed</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_APPLET_PROPERTY_NOT_SUPPORTED"/></term><description>The property requested is not supported</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_APPLET_GUID"/></term><description>The applet id is invalid</description></item>
        /// </list>
        /// </exception>
        public void GetAppletProperty(string AppId, string Property, out string Value)
        {
            uint ret;

            uint propertyLen = DEFUALT_APPLET_PROPERTY_LEN;

            byte[] OutBuf = new byte[(int) (propertyLen + 1) * 2];
            
            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (AppId == null)
            {
                throw new ArgumentNullException("AppId");
            }

            if (Property == null)
            {
                throw new ArgumentNullException("Property");
            }

            Value = "";


            byte[] propertyBuf = System.Text.Encoding.Unicode.GetBytes(Property);
            byte[] Null = { 0x00, 0x00 };
            byte[] InBuf = new byte[propertyBuf.Length + Null.Length];

            Array.Copy(propertyBuf,0, InBuf,0, propertyBuf.Length);
            Array.Copy(Null, 0, InBuf, propertyBuf.Length, Null.Length);

            IntPtr tmpInBuf = Marshal.AllocHGlobal(InBuf.Length);

            Marshal.Copy(InBuf, 0, tmpInBuf, InBuf.Length);
            
            

            IntPtr tmpOutBuf = Marshal.AllocHGlobal(OutBuf.Length);



            JhiWrapper.JVM_COMM_BUFFER comm;
            comm.TxBuf.buffer = tmpInBuf;
            comm.TxBuf.length = (uint) Property.Length;
            comm.RxBuf.buffer = tmpOutBuf;
            comm.RxBuf.length = propertyLen;


            ret = JhiWrapper.JHI_GetAppletProperty(_handle, RemoveSeperators(AppId), ref comm);

            if (Enum.IsDefined(typeof(JHI_ERROR_CODE), ret))
            {
                JHI_ERROR_CODE _ret = (JHI_ERROR_CODE)ret;
                if (_ret == JHI_ERROR_CODE.JHI_INSUFFICIENT_BUFFER)
                {
                    Marshal.FreeHGlobal(tmpOutBuf);
                    propertyLen = comm.RxBuf.length;
                    OutBuf = new byte[(int)(propertyLen + 1) * 2];
                    tmpOutBuf = Marshal.AllocHGlobal(OutBuf.Length);
                    comm.RxBuf.buffer = tmpOutBuf;

                    ret = JhiWrapper.JHI_GetAppletProperty(_handle, RemoveSeperators(AppId), ref comm);
                }
            }


            if (ret != 0)
            {
                Marshal.FreeHGlobal(tmpInBuf);
                Marshal.FreeHGlobal(tmpOutBuf);
                throw new JhiException("JHI_GetAppletProperty() failed", ret);
            }

            if (OutBuf.Length > 0)
                Marshal.Copy(tmpOutBuf, OutBuf, 0, OutBuf.Length);

            Value = System.Text.Encoding.Unicode.GetString(OutBuf, 0,(int) comm.RxBuf.length * 2);

            Marshal.FreeHGlobal(tmpInBuf);
            Marshal.FreeHGlobal(tmpOutBuf);
        }


        /// <summary>
        /// Create a session of an installed applet. 
        /// a session handle is retuned by SessionHandle.
        /// </summary>
        /// <param name="AppId">Applet ID</param>
        /// <param name="Session">Session Handle</param>
        /// <param name="initBuffer">Initialization data passed to the applet onInit function</param>
        /// <param name="flags">session flags used for creation</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_APPLET_GUID"/></term><description>The applet id is invalid</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_APPLET_FATAL"/></term><description>Returned when the applet session has crashed</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_BUFFER_SIZE"/></term><description>Used a initBuffer that is larger than <see cref="Jhi.JHI_BUFFER_MAX"/></description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_APPLET_NOT_INSTALLED"/></term><description>The applet is not installed</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_MAX_SESSIONS_REACHED"/></term><description>Reached the limit of sessions in FW</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SHARED_SESSION_NOT_SUPPORTED"/></term><description>the applet does not support shared sessions</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_MAX_SHARED_SESSION_REACHED"/></term><description>Reached the limit of handles to the shared session</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_ONLY_SINGLE_INSTANCE_ALLOWED"/></term><description>Returned if more than single instance was opened</description></item>
        /// </list>
        /// </exception>
        public void CreateSession(string AppId, JHI_SESSION_FLAGS flags, byte[] initBuffer, out JhiSession Session)
        {
            uint ret;
            bool memAlocated = false;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (AppId == null)
            {
                throw new ArgumentNullException("AppId");
            }

            Session = new JhiSession();

            JhiWrapper.DATA_BUFFER init_buffer = new JhiWrapper.DATA_BUFFER();
            if (initBuffer == null)
            {
                init_buffer.length = 0;
            }
            else
            {
                init_buffer.length = (uint) initBuffer.Length;
                init_buffer.buffer = Marshal.AllocHGlobal(initBuffer.Length);
                memAlocated = true;

                if (initBuffer.Length > 0)
                    Marshal.Copy(initBuffer, 0, init_buffer.buffer, initBuffer.Length);
            }

            ret = JhiWrapper.JHI_CreateSession(_handle, RemoveSeperators(AppId), Convert.ToUInt32(flags), ref init_buffer, ref Session.SessionHandle);

            if (memAlocated)
                Marshal.FreeHGlobal(init_buffer.buffer);

            if (ret != 0)
                throw new JhiException("JHI_CreateSession() failed", ret);
        }

        /// <summary>
        /// Close an applet session
        /// </summary>
        /// <param name="Session">Session Handle</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_SESSION_HANDLE"/></term><description>The session handle is not valid</description></item>
        /// </list>
        /// </exception>
        public void CloseSession(JhiSession Session)
        {
            uint ret;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (Session == null)
            {
                throw new ArgumentNullException("Session");
            }

            ret = JhiWrapper.JHI_CloseSession(_handle, ref Session.SessionHandle);
            if (ret != 0)
                throw new JhiException("JHI_CloseSession() failed", ret);
        }

        public void ForceCloseSession(JhiSession Session)
        {
            uint ret;
            if (_handle == null)
                throw new JhiException("JHI is not initalized");
            if (Session == null)
            {
                throw new ArgumentNullException("Session");
            }
            ret = JhiWrapper.JHI_ForceCloseSession(_handle, ref Session.SessionHandle);
            if (ret != 0)
                throw new JhiException("JHI_ForceCloseSession() failed", ret);
        }
        /// <summary>
        /// Get the number of existing sessions of an applet 
        /// </summary>
        /// <param name="AppId">Applet ID</param>
        /// <param name="SessionsCount">The number of sessions</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_APPLET_GUID"/></term><description>The applet id is invalid</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_APPLET_NOT_INSTALLED"/></term><description>The applet is not installed and does not exist in JHI repository</description></item>
        /// </list>
        /// </exception>
        public void GetSessionsCount(string AppId, out UInt32 SessionsCount)
        {
            uint ret;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (AppId == null)
            {
                throw new ArgumentNullException("AppId");
            }

            SessionsCount = 0;

            ret = JhiWrapper.JHI_GetSessionsCount(_handle, RemoveSeperators(AppId), ref SessionsCount);
            if (ret != 0)
                throw new JhiException("JHI_GetSessionsCount() failed", ret);
        }

        /// <summary>
        /// Get information of a given session
        /// </summary>
        /// <param name="Session">Session Handle</param>
        /// <param name="SessionInfo">The session info</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_SESSION_HANDLE"/></term><description>The session handle is not valid</description></item>
        /// </list>
        /// </exception>
        public void GetSessionInfo(JhiSession Session, out JHI_SESSION_INFO SessionInfo)
        {
            uint ret;
            JhiWrapper.JHI_I_SESSION_INFO info = new JhiWrapper.JHI_I_SESSION_INFO();

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (Session == null)
            {
                throw new ArgumentNullException("Session");
            }

            ret = JhiWrapper.JHI_GetSessionInfo(_handle, Session.SessionHandle, ref info);
            if (ret != 0)
                throw new JhiException("JHI_GetSessionInfo() failed", ret);

            SessionInfo = new JHI_SESSION_INFO();
            SessionInfo.state = info.state;
            SessionInfo.flags = (JHI_SESSION_FLAGS) info.flags;
        }

        /// <summary>
        /// Registers registration of one callback function to receive events from a given session
        /// </summary>
        /// <param name="Session">Session Handle</param>
        /// <param name="CallbackFunction">The Callback function</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SESSION_ALREADY_REGSITERED"/></term><description>The session is already registered for events</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_SESSION_HANDLE"/></term><description>The session handle is not valid</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_EVENTS_NOT_SUPPORTED "/></term><description>Events are not supported for this type of session</description></item>
        /// </list>
        /// </exception>
        public void RegisterEvents(JhiSession Session, Intel.Dal.JHI_CallbackFunction CallbackFunction)
        {
            uint ret;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (Session == null)
            {
                throw new ArgumentNullException("Session");
            }

            if (CallbackFunction == null)
            {
                throw new ArgumentNullException("CallbackFunction");
            }

            //ret = JhiWrapper.JHI_RegisterEvents(_handle, SessionHandle, CallbackFunction);
            ret = JhiWrapper.JHI_RegisterEvents(_handle, Session.SessionHandle, /*CallbackFunc*/ JhiMainCallback);

            if (ret == 0)
            {
                Session.callback = CallbackFunction;
                lock (_eventsLock)
                {
                    _eventList.Add(Session);
                }
            }
            else
            {
                throw new JhiException("JHI_RegisterEvents() failed", ret);
            }
        }


        /// <summary>
        ///  Remove registration of events from a given session
        /// </summary>
        /// <param name="Session">Session Handle</param>
        /// <exception cref="System.ArgumentNullException">Thrown when passed a null to one of the method arguments</exception>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SESSION_NOT_REGISTERED"/></term><description>The session wasn’t registered for events</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INVALID_SESSION_HANDLE"/></term><description>The session handle is not valid</description></item>
        /// </list>
        /// </exception>
        public void UnRegisterEvents(JhiSession Session)
        {
            uint ret;

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            if (Session == null)
                throw new ArgumentNullException("Session");

            ret = JhiWrapper.JHI_UnRegisterEvents(_handle, Session.SessionHandle);

            if (ret == 0)
            {
                lock (_eventsLock)
                {
                    _eventList.Remove(Session);
                }
            }
            else throw new JhiException("JHI_UnRegisterEvents() failed", ret);
        }

        /// <summary>
        /// This function is used in order to retrieve global information of JHI such as JHI versions and the FW versions,
        /// see the JHI_VERSION_INFO for the specific info retuned.
        /// </summary>
        /// <param name="VersionInfo">a struct that holds the versions.</param>
        /// <exception cref="JhiException">
        /// Thrown when the operation failed. The specific error resides within the class JhiRet member
        /// 
        /// <list type="table">
        /// <listheader>
        /// <term>Possible error codes</term>
        /// <description>Description</description>
        /// </listheader>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_INTERNAL_ERROR"/></term><description>Returned if the calling functions return an internal error</description></item>
        /// <item><term><see cref="JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE"/></term><description>Returned when there is no connection to the JHI service</description></item>
        /// </list>
        /// </exception>
        public void GetVersionInfo(out JHI_VERSION_INFO VersionInfo)
        {
            uint ret;
            JhiWrapper.JHI_I_VERSION_INFO info = new JhiWrapper.JHI_I_VERSION_INFO();

            if (_handle == null)
                throw new JhiException("JHI is not initalized");

            ret = JhiWrapper.JHI_GetVersionInfo(_handle, ref info);

            if (ret != 0)
            {
                throw new JhiException("JHI_GetVersionInfo() failed", ret);
            }

            VersionInfo = new JHI_VERSION_INFO();

            VersionInfo.jhi_version = info.jhi_version;
            VersionInfo.fw_version = info.fw_version;
            VersionInfo.comm_type = info.comm_type;
            VersionInfo.platform_id = info.platform_id;
        }

        #endregion

        /// <summary>
        ///  This callback function is used in order to recieve a event from the native jhi.dll and
        ///  translate it to a managed callback function.
        /// </summary>
        /// <param name="SessionHandle">the session that invoked the event</param>
        /// <param name="Event">the event data</param>
        private void CallbackFunc(IntPtr SessionHandle, JHI_I_EVENT_DATA Event)
        {
            //find the session in the event list and invoke the callback
            foreach (JhiSession Session in _eventList)
            {
                if (IntPtr.Equals(Session.SessionHandle, SessionHandle))
                {
                    JHI_EVENT_DATA eventData;
                    eventData.datalen = Event.datalen;
                    eventData.dataType = Event.dataType;

                    eventData.dataBuffer = new byte[Event.datalen];

                    if (eventData.dataBuffer.Length > 0)
                        Marshal.Copy(Event.data, eventData.dataBuffer, 0, eventData.dataBuffer.Length);

                    Session.callback(Session, eventData);
                    break;
                }
            }
        }
        
        
        /// <summary>
        /// Removes all the '-' characters from an applet id.
        /// </summary>
        /// <param name="AppId">Applet ID</param>
        /// <returns>Applet ID without '-' chars</returns>
        private static string RemoveSeperators(string AppId)
        {
            return AppId.Replace("-", "");
        }

    }
}

// Restore missing documentation warnings
#pragma warning restore 1591
