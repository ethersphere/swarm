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
    /// this enum lists the types of data received by JHI event
    /// </summary>
    public enum JHI_EVENT_DATA_TYPE
    {
        /// <summary>
        ///  the event raised by an applet session
        /// </summary>
        JHI_DATA_FROM_APPLET = 0,

        /// <summary>
        /// the event raised by JHI service
        /// </summary>
        JHI_DATA_FROM_SERVICE = 1
    }

    /// <summary>
    /// this struct repersents the data received upon a JHI event 
    /// </summary>
    public struct JHI_EVENT_DATA
    {
        /// <summary>
        /// byte length of the event data
        /// </summary>
        public UInt32 datalen;

        /// <summary>
        /// the buffer that contains the event data
        /// </summary>
        public byte [] dataBuffer;

        /// <summary>
        /// the event type
        /// </summary>
        public JHI_EVENT_DATA_TYPE dataType;
    }

    /// <summary>
    /// This is the format for a callback function that is used in order to 
    /// receive session events.
    /// </summary>
    /// <param name="SessionHandle">a handle for the session raised the event</param>
    /// <param name="event_data">the event data</param>
    public delegate void JHI_CallbackFunction(JhiSession SessionHandle, JHI_EVENT_DATA event_data);

    /// <summary>
    /// this class is used as a handle for an applet session
    /// </summary>
    public class JhiSession
    {
        internal IntPtr SessionHandle;
        internal JHI_CallbackFunction callback;

        /// <summary>
        /// defualt constructor for this class
        /// </summary>
        public JhiSession()
        {
            SessionHandle = new IntPtr();
            callback = null;
        }
    }
}
