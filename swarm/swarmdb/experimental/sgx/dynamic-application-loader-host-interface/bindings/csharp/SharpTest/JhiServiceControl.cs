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
using System.Management;
using System.ServiceProcess;
using System.Text;

namespace JhiServiceControl
{
    public static class JhiServiceController
    {
        static string JhiServiceName = "jhi_service";

        #region fromDNA (edited)


        /// <summary>
        /// Resets the Jhi Service
        /// </summary>
        public static bool ResetJhiService()
        {
            return StopJhiService() && StartJhiService();
        }

        /// <summary>
        /// Stops the JHi service
        /// </summary>
        public static bool StopJhiService()
        {
            try
            {
                ServiceController service = new ServiceController(JhiServiceName);
                service.Stop();
                service.WaitForStatus(ServiceControllerStatus.Stopped);
                service.Refresh();
                if (service.Status == ServiceControllerStatus.Stopped)
                {
                    service.Close();
                    System.Threading.Thread.Sleep(1000);
                    return true;
                }
                return false;
            }
            catch (Exception ex)
            {
                Console.WriteLine("Failed to stop the service " + ex.Message);
                return false;
            }
            
        }

        /// <summary>
        /// Starts the JHI service
        /// </summary>
        public static bool StartJhiService()
        {
            try
            {
                ServiceController service = new ServiceController(JhiServiceName);
                service.Start();
                service.WaitForStatus(ServiceControllerStatus.Running);
                service.Refresh();
                if (service.Status == ServiceControllerStatus.Running)
                {
                    service.Close();
                    System.Threading.Thread.Sleep(1000);
                    return true;
                }
                return false;
            }
            catch (Exception ex)
            {
                Console.WriteLine("Failed to start the service " + ex.Message);
                return false;
            }

        }

        /// <summary>
        /// The method returns the JHI status
        /// </summary>
        public static ServiceControllerStatus GetJhiStatus()
        {
            ServiceController service = new ServiceController(JhiServiceName);
            ServiceControllerStatus tmp = service.Status;
            service.Close();
            return tmp;
        }

        /// <summary>
        /// Returns true if JHI status equals to the expected status and false otherwise
        /// </summary>
        public static bool CheckJhiStatus(ServiceControllerStatus expectedStatus)
        {
            ServiceControllerStatus jhiStatus = GetJhiStatus();
            if (!expectedStatus.Equals(jhiStatus))
            {
                return false;
            }
            return true;
        }

        #endregion

    }
}
