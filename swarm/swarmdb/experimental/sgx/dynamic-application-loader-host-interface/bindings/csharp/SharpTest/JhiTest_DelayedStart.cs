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
using Intel.Dal;
using JhiServiceControl;

namespace SharpTest
{
    public static class JhiTest_DelayedStart
    {
        public static Jhi jhi;

        public static bool test()
        {
            try
            {
                Jhi.DisableDllValidation = true;

                //JHI_VERSION_INFO VersionInfo = new JHI_VERSION_INFO();
                //jhi.GetVersionInfo(out VersionInfo);
                //Console.WriteLine(VersionInfo.fw_version);

                if (JhiServiceController.CheckJhiStatus(System.ServiceProcess.ServiceControllerStatus.Running))
                {
                    Console.WriteLine("JHI Status = " + JhiServiceController.GetJhiStatus());
                    Console.WriteLine("Stopping service...");
                    if (JhiServiceController.StopJhiService())
                    {
                        Console.WriteLine("Success!");
                    }
                    else
                    {
                        Console.WriteLine("Failed to stop service!");
                        return false;
                    }
                }
                Console.WriteLine("JHI Status = " + JhiServiceController.GetJhiStatus());
                Console.WriteLine("Starting service...");
                if (JhiServiceController.StartJhiService())
                {
                    Console.WriteLine("Success!");
                }
                else
                {
                    Console.WriteLine("Failed to start service!");
                    return false;
                }

                Console.WriteLine("JHI Status = " + JhiServiceController.GetJhiStatus());
                if (!JhiServiceController.CheckJhiStatus(System.ServiceProcess.ServiceControllerStatus.Running))
                {
                    Console.WriteLine("JHI isn't running");
                    return false;
                }


                Console.WriteLine("Attempting to init... ");
                jhi = Jhi.Instance;
                Console.WriteLine("Init succeeded.");
            }
            catch (JhiException e)
            {
                Console.WriteLine(e.Message);
                return false;
            }
            catch (Exception e)
            {
                Console.WriteLine(e.Message);
                Console.WriteLine(e.InnerException);
                return false;
            }
            return true;
        }
    }
}
