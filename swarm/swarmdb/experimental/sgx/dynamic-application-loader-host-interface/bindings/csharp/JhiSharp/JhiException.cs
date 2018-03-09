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

﻿using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace Intel.Dal
{
    /// <summary>
    /// this exception is raised by JHI when a call to a JHI API failes
    /// </summary>
    [Serializable]
    public class JhiException : Exception
    {
        private JHI_ERROR_CODE? _ret;

        /// <summary>
        /// the error code
        /// </summary>
        public JHI_ERROR_CODE JhiRet 
        { 
            get
            {
                return _ret.Value;
            }
        }

        private static Dictionary<JHI_3_TEE_ERROR_CODES, JHI_ERROR_CODE> TeeErrorToJhiErrorMap = new Dictionary<JHI_3_TEE_ERROR_CODES, JHI_ERROR_CODE>
        {
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_INTERNAL_ERROR, JHI_ERROR_CODE.JHI_INTERNAL_ERROR },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_INVALID_PARAMS, JHI_ERROR_CODE.JHI_INVALID_PARAMS },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_INVALID_HANDLE, JHI_ERROR_CODE.JHI_INVALID_HANDLE },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_INVALID_UUID, JHI_ERROR_CODE.JHI_INVALID_APPLET_GUID },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_NO_FW_CONNECTION, JHI_ERROR_CODE.JHI_NO_CONNECTION_TO_FIRMWARE },
            //{ JHI_3_TEE_ERROR_CODES.TEE_STATUS_NOT_SUPPORTED, JHI_ERROR_CODE.? },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_SERVICE_UNAVAILABLE, JHI_ERROR_CODE.JHI_SERVICE_UNAVAILABLE },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_REGISTRY_ERROR, JHI_ERROR_CODE.JHI_ERROR_REGISTRY },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_REPOSITORY_ERROR, JHI_ERROR_CODE.JHI_ERROR_REPOSITORY_NOT_FOUND },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_SPOOLER_MISSING, JHI_ERROR_CODE.JHI_SPOOLER_NOT_FOUND },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_SPOOLER_INVALID, JHI_ERROR_CODE.JHI_INVALID_SPOOLER },
            //{ JHI_3_TEE_ERROR_CODES.TEE_STATUS_MISSING_PLUGIN, JHI_ERROR_CODE.? },
            //{ JHI_3_TEE_ERROR_CODES.TEE_STATUS_PLUGIN_VERIFY_FAILED, JHI_ERROR_CODE.? },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_INVALID_PACKAGE, JHI_ERROR_CODE.JHI_INVALID_PACKAGE_FORMAT },
            //{ JHI_3_TEE_ERROR_CODES.TEE_STATUS_PACKAGE_AUTHENTICATION_FAILURE, JHI_ERROR_CODE.? },
            //{ JHI_3_TEE_ERROR_CODES.TEE_STATUS_MAX_SVLS_REACHED, JHI_ERROR_CODE.? },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_CMD_FAILURE_SESSIONS_EXISTS, JHI_ERROR_CODE.JHI_INSTALL_FAILURE_SESSIONS_EXISTS },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_CMD_FAILURE, JHI_ERROR_CODE.JHI_INSTALL_FAILED },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_MAX_TAS_REACHED, JHI_ERROR_CODE.JHI_MAX_INSTALLED_APPLETS_REACHED },
            //{ JHI_3_TEE_ERROR_CODES.TEE_STATUS_MISSING_ACCESS_CONTROL, JHI_ERROR_CODE.? },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_TA_DOES_NOT_EXIST, JHI_ERROR_CODE.JHI_APPLET_NOT_INSTALLED },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_SVL_CHECK_FAIL, JHI_ERROR_CODE.JHI_SVL_CHECK_FAIL },
            { JHI_3_TEE_ERROR_CODES.TEE_STATUS_IDENTICAL_PACKAGE, JHI_ERROR_CODE.JHI_INSTALL_FAILED }
        };

        /// <summary>
        /// defualt constructor
        /// </summary>
        public JhiException() { }

        /// <summary>
        /// constructor overload
        /// </summary>
        /// <param name="message">exception message</param>
        public JhiException(string message) : base(message) { }

        /// <summary>
        /// constructor overload
        /// </summary>
        /// <param name="message">exception message</param>
        /// <param name="ret">the error code</param>
        public JhiException(string message, uint ret) : base(message) 
        {
            _ret = GetExceptionErrorCode(ret);
        }

        private JHI_ERROR_CODE GetExceptionErrorCode(uint ret)
        {
            if (Enum.IsDefined(typeof(JHI_ERROR_CODE), ret))
                return (JHI_ERROR_CODE)ret;
            
            if (Enum.IsDefined(typeof(JHI_3_TEE_ERROR_CODES), (int)ret) && TeeErrorToJhiErrorMap.ContainsKey((JHI_3_TEE_ERROR_CODES)ret))
            {
                return TeeErrorToJhiErrorMap[(JHI_3_TEE_ERROR_CODES)ret];
            }
            return JHI_ERROR_CODE.JHI_UNKOWN_ERROR_CODE;
        }
           
        



        /// <summary>
        /// constructor overload
        /// </summary>
        /// <param name="message">exception message</param>
        /// <param name="inner">inner exception</param>
        public JhiException(string message, Exception inner) : base(message, inner) { }
        
        /// <summary>
        /// constructor overload
        /// </summary>
        /// <param name="message">exception message</param>
        /// <param name="ret">the error code</param>
        /// <param name="inner">inner exception</param>
        public JhiException(string message, uint ret, Exception inner) : base(message, inner) 
        {
            if (Enum.IsDefined(typeof(JHI_ERROR_CODE), ret))
                _ret = (JHI_ERROR_CODE)ret;
            else
                _ret = JHI_ERROR_CODE.JHI_INTERNAL_ERROR;
        }

        /// <summary>
        /// constructor overload
        /// </summary>
        /// <param name="info"></param>
        /// <param name="context"></param>
        protected JhiException(
          System.Runtime.Serialization.SerializationInfo info,
          System.Runtime.Serialization.StreamingContext context)
            : base(info, context) { }

        /// <summary>
        /// convert the exception to string
        /// </summary>
        /// <returns>the string</returns>
        public override string ToString()
        {
            if (_ret.HasValue)
            {
                return string.Format("{0} [{1}]", base.ToString(), _ret);
            }
            
            return base.ToString();
        }
    }

    /// <summary>
    /// this exception is raised when a response buffer sent to a session
    /// is insufficent.
    /// </summary>
    [Serializable]
    public class JhiInsufficientBufferException : JhiException
    {
        private uint? _required_size = null;

        /// <summary>
        /// the buffer size required
        /// </summary>
        public uint Required_size
        {
            get
            {
                return _required_size.Value;
            }
        }

        /// <summary>
        /// a defualt constructor for this class 
        /// </summary>
        /// <param name="message">the message</param>
        /// <param name="ret">the error code value</param>
        /// <param name="required_size">the size required for the response buffer</param>
        public JhiInsufficientBufferException(string message, uint ret, uint required_size)
            : base(message, ret)
        { 
            _required_size = required_size; 
        }
    }
}
