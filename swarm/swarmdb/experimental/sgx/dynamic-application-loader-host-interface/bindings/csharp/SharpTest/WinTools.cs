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
using System.Threading;
using Microsoft.WDTF;

namespace SharpTest
{
    public static class WinTools
    {
        /// <summary>
        /// 
        /// </summary>
        /// <param name="delay">The time that the system will be online before the CS duration (milliseconds)</param>
        /// <param name="duration">The time that the system will be in CS (milliseconds)</param>
        /// <param name="block">Determines whether to be blocked until system resumes.</param>
        /// <returns></returns>
        public static bool setConnectedStandby(int delay, int duration, bool block)
        {
            return setConnectedStandby(delay, duration, 1, block);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="delay">The time that the system will be online before and between the CS durations (milliseconds)</param>
        /// <param name="duration">The time that the system will be in CS in every iteration (milliseconds)</param>
        /// <param name="iterations">The number of times to set a CS state.</param>
        /// <param name="block">Determines whether to be blocked until system resumes.</param>
        /// <returns></returns>
        public static bool setConnectedStandby(int delay, int duration, int iterations, bool block)
        {
            try
            {
                if (delay < 0 || duration < 1 || iterations < 1)
                {
                    return false;
                }
                Thread thread = new Thread( () => _setConnectedStandby(delay, duration, iterations));
                thread.IsBackground = true;
                thread.Start();
                if (block)
                {
                    thread.Join();
                }
                return true;
            }
            catch (Exception)
            {
                return false;
            }
        }

        private static void _setConnectedStandby(int delay, int duration, int iterations)
        {
            try
            {
                for (int i = 0; i < iterations; i++)
                {
                    Thread.Sleep(delay);
                    var wdtf = new WDTF2();
                    var system = (IWDTFSystemAction2)wdtf.SystemDepot.ThisSystem.GetInterface("System");
                    system.ConnectedStandby(duration);
                }
            }
            catch (Exception)
            {
            }
        }
    }
}
