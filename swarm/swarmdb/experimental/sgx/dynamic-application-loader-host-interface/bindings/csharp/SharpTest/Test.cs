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
using System.Collections;
using System.Linq;
using System.Text;
using Intel.Dal;
using System.IO;
using System.Threading;
using System.Windows.Forms;
using JhiServiceControl;
using System.Security.Principal;

namespace SharpTest
{
    public class JhiTests
    {
        #region Private members
        private static Jhi jhi;
        private static string echoAppID = "d1de41d82b844feaa7fa1e4322f15dee";
        private static string echoAppPath = Directory.GetCurrentDirectory() + "/echo.dalp";
        private static string eventServiceAppID = "a525599fc5214aae9f952f268fa54416";
        private static string eventServiceAppPath = Directory.GetCurrentDirectory() + "/eventservice.dalp";
        private static string eventServiceWithTimeoutAppID = "a525599fc5214aae9f952f268fa54417";
        private static string eventServiceWithTimeoutAppPath = Directory.GetCurrentDirectory() + "/EventServiceWithTimeout.dalp";
        private static bool invoked = false;
        private static bool isAdmin = IsAdministrator();

        private static bool IsAdministrator()
        {
            var identity = WindowsIdentity.GetCurrent();
            var principal = new WindowsPrincipal(identity);
            return principal.IsInRole(WindowsBuiltInRole.Administrator);
        }

        private static void callbackFun(JhiSession SessionHandle, JHI_EVENT_DATA event_data)
        {
            Console.WriteLine("Got event!");
            invoked = true;
        }

        private static void stopJHISporadically()
        {
            Random random = new Random();
            int randomNumber;

            for (int i = 0; i < 15; i++)
            {
                randomNumber = random.Next(0, 500);
                Thread.Sleep(randomNumber);
                JhiServiceController.ResetJhiService();
            }
        }

        private static bool ArraysEqual<T>(T[] a1, T[] a2)
        {
            if (ReferenceEquals(a1, a2))
                return true;

            if (a1 == null || a2 == null)
                return false;

            if (a1.Length != a2.Length)
                return false;

            EqualityComparer<T> comparer = EqualityComparer<T>.Default;
            for (int i = 0; i < a1.Length; i++)
            {
                if (!comparer.Equals(a1[i], a2[i])) return false;
            }
            return true;
        }
        #endregion

        #region Public members

        public enum JhiTestResult
        {
            Succeeded,
            Skipped,
            Failed
        }

        public static bool init()
        {
            Jhi.DisableDllValidation = true;
            jhi = Jhi.Instance;
            return true;
        }

        public static JhiTestResult GetFWVersion()
        {
            try
            {
                JHI_VERSION_INFO VersionInfo;
                jhi.GetVersionInfo(out VersionInfo);
                Console.WriteLine("FW Version: " + VersionInfo.fw_version);
                Console.WriteLine("JHI Version: " + VersionInfo.jhi_version);
                Console.WriteLine("Platform ID: " + VersionInfo.platform_id);
                return JhiTestResult.Succeeded;
            }
            catch (Exception e)
            {
                Console.WriteLine(e.Message);
                return JhiTestResult.Failed;
            }
        }

        public static JhiTestResult bool_To_JhiTestResult(bool result)
        {
            if (result)
            {
                return JhiTestResult.Succeeded;
            }
            return JhiTestResult.Failed;
        }

        public static JhiTestResult DelayedStart()
        {
            if (!isAdmin)
            {
                Console.WriteLine("Admin privileges requierd for this test! skipping...");
                return JhiTestResult.Skipped;
            }
            return bool_To_JhiTestResult(JhiTest_DelayedStart.test());
        }


        public static JhiTestResult SendAndReceive()
        {
            Console.Write("Installing App " + echoAppID + "...  ");
            jhi.Install(echoAppID, echoAppPath);
            jhi.Uninstall(echoAppID);
            jhi.Install(echoAppID, echoAppPath);
            Console.WriteLine("Success.");

            Console.Write("Creating session...  ");
            JhiSession echoSession;
            jhi.CreateSession(echoAppID, JHI_SESSION_FLAGS.None, null, out echoSession);
            Console.WriteLine("Success.");

            Console.Write("Starting send and receive...  ");
            int output;
            Byte[] x = new Byte[]{3,12,3};
            Byte[] y = new Byte[3];
            jhi.SendAndRecv2(echoSession, 1, x,ref y, out output);
            bool result = ArraysEqual(x, y);
            Console.WriteLine("Output matches! - success.");
            jhi.CloseSession(echoSession);
            jhi.Uninstall(echoAppID);
            return bool_To_JhiTestResult(result);
        }


        public static JhiTestResult StopWhileSendAndReceive()
        {
            if (!isAdmin)
            {
                Console.WriteLine("Admin privileges requierd for this test! skipping...");
                return JhiTestResult.Skipped;
            }
            Console.Write("Installing App " + echoAppID + "...  ");
            jhi.Install(echoAppID, echoAppPath);
            Console.WriteLine("Success.");

            Console.Write("Creating session...  ");
            JhiSession echoSession;
            jhi.CreateSession(echoAppID, JHI_SESSION_FLAGS.None, null, out echoSession);
            Console.WriteLine("Success.");

            int output;
            Byte[] x = new Byte[] { 3, 12, 3 };
            Byte[] y = new Byte[3];

            // setting a thread to stop the service from time to time
            Thread jhi_stopper = new Thread(stopJHISporadically);
            jhi_stopper.Start();

            Console.Write("Starting send and receive...  ");
            bool session_created;
			
			for (int i = 0; i < 500; i++)
            {
                try
                {
                    jhi.SendAndRecv2(echoSession, 1, x, ref y, out output);
                    Console.WriteLine("S&R succeeded.");
                }
                catch (JhiException ex)
                {
                    Console.WriteLine("S&R return code = " + ex.JhiRet + ", " + ex.Message + ", " + ex.InnerException);
                    session_created = false;
                    while (!session_created)
                    {
                        session_created = false;
                        try
                        {
                            Console.Write("Trying to create session...  ");
                            jhi.CreateSession(echoAppID, JHI_SESSION_FLAGS.None, null, out echoSession);
                            Console.WriteLine("Success.");
                            session_created = true;

                        }
                        catch (Exception)
                        {
                            Console.WriteLine("failed.");
                        }
                    }
                }
            }
            jhi.CloseSession(echoSession);
            jhi.Uninstall(echoAppID);
            return JhiTestResult.Succeeded;
        }


        public static JhiTestResult eventDuringSleepTest()
        {
            invoked = false;
            Console.Write("Installing App " + eventServiceWithTimeoutAppID + "...  ");
            jhi.Install(eventServiceWithTimeoutAppID, eventServiceWithTimeoutAppPath);
            Console.WriteLine("Success. " + DateTime.Now);

            Console.Write("Creating session...  ");
            JhiSession eventsSession;
            jhi.CreateSession(eventServiceWithTimeoutAppID, JHI_SESSION_FLAGS.None, null, out eventsSession);
            Console.WriteLine("Success. " + DateTime.Now);

            Console.Write("Registering for events...  ");
            jhi.RegisterEvents(eventsSession, callbackFun);
            Console.WriteLine("Success. " + DateTime.Now);

            #region event receice during sleep

            //Console.WriteLine("Success. " + DateTime.Now);

            Console.Write("Starting send and receive...  " + DateTime.Now);
            int output;
            Byte[] x = new Byte[] { 15 }; // used as seconds to before event
            Byte[] y = new Byte[3];
            jhi.SendAndRecv2(eventsSession, 20, x, ref y, out output);
            Console.WriteLine("Success. " + DateTime.Now);
            Console.WriteLine("An event should be raised in 15 seconds " + DateTime.Now);

            Console.Write("Setting PC to go to CS for 30 seconds. " + DateTime.Now);
            WinTools.setConnectedStandby(0, 30000, true);


            for (int i = 0; i < 150; i++)
            {
                Thread.Sleep(100);
                if (invoked)
                {
                    Console.WriteLine("Event raised! - success. " + DateTime.Now);
                    break;
                }
            }

            if (!invoked)
            {
                Console.WriteLine("Event not received!");
                return JhiTestResult.Failed;
            }
            invoked = false;

            #endregion

            #region event receice after sleep

            //Console.WriteLine("Success. " + DateTime.Now);

            Console.Write("Starting send and receive...   " + DateTime.Now);
            int output2;
            Byte[] x2 = new Byte[] { 35 }; // used as seconds to before event
            Byte[] y2 = new Byte[3];
            jhi.SendAndRecv2(eventsSession, 20, x2, ref y2, out output2);
            Console.WriteLine("Success. " + DateTime.Now);
            Console.WriteLine("An event should be raised in 35 seconds. " + DateTime.Now);

            Console.Write("Setting PC to go to CS for 15 seconds. " + DateTime.Now);
            WinTools.setConnectedStandby(0, 30000, true);


            #endregion



            for (int i = 0; i < 150; i++)
            {
                Thread.Sleep(100);
                if (invoked)
                {
                    Console.WriteLine("Event raised! - success. " + DateTime.Now);
                    break;
                }
            }
            return bool_To_JhiTestResult(invoked);
        }

        public static JhiTestResult SendAndReceiveDuringSleepTest()
        {
            //should be manual test
            int output = -1;
            Console.Write("Installing App " + eventServiceWithTimeoutAppID + "...  ");
            jhi.Install(eventServiceWithTimeoutAppID, eventServiceWithTimeoutAppPath);
            Console.WriteLine("Success. " + DateTime.Now);

            Console.Write("Creating session...  ");
            JhiSession eventsSession;
            jhi.CreateSession(eventServiceWithTimeoutAppID, JHI_SESSION_FLAGS.None, null, out eventsSession);
            Console.WriteLine("Success. " + DateTime.Now);


            #region sendAndReceice finish during sleep

            Console.Write("Setting PC to go to CS for 25 seconds in 2 seconds from now. " + DateTime.Now);
            WinTools.setConnectedStandby(2000, 25000, false);


            Console.Write("Starting send and receive for 10 seconds...  " + DateTime.Now);
            output = -1;
            Byte[] x = new Byte[] { 10 }; // +/- seconds
            Byte[] y = new Byte[3];
            jhi.SendAndRecv2(eventsSession, 40, x, ref y, out output);
            Console.WriteLine("output = " + output + " " + DateTime.Now);

            if (output!=0)
            {
                return JhiTestResult.Failed;
            }

            #endregion

            #region sendAndReceice finish after sleep

            Console.Write("Setting PC to go to CS for 10 seconds in 2 seconds from now. " + DateTime.Now);
            WinTools.setConnectedStandby(2000, 10000, false);


            Console.Write("Starting send and receive for 35 seconds...  " + DateTime.Now);
            output = -1;
            Byte[] x2 = new Byte[] { 35 }; // +/- seconds
            Byte[] y2 = new Byte[3];
            jhi.SendAndRecv2(eventsSession, 40, x2, ref y2, out output);
            Console.WriteLine("output = " + output + " " + DateTime.Now);

            if (output != 0)
            {
                return JhiTestResult.Failed;
            }

            #endregion

           
            return JhiTestResult.Succeeded;
        }

        public static JhiTestResult eventsTest()
        {
            invoked = false;
            Console.Write("Installing App " + eventServiceAppID + "...  ");
            jhi.Install(eventServiceAppID, eventServiceAppPath);
            Console.WriteLine("Success.");

            Console.Write("Creating session...  ");
            JhiSession eventsSession;
            jhi.CreateSession(eventServiceAppID, JHI_SESSION_FLAGS.None, null, out eventsSession);
            Console.WriteLine("Success.");

            try // testing unregister events
            {
                Console.WriteLine("Trying to unregister, expecting JHI_SESSION_NOT_REGISTERED...");
                jhi.UnRegisterEvents(eventsSession);
            }
            catch (JhiException e)
            {
                if (!(e.JhiRet == JHI_ERROR_CODE.JHI_SESSION_NOT_REGISTERED))
                {
                    Console.WriteLine("Unexpected exception received: " + e.JhiRet);
                    throw;
                }
            }
            Console.WriteLine("Expected exception received.");

            Console.Write("Registering for events...  ");
            jhi.RegisterEvents(eventsSession, callbackFun);
            Console.WriteLine("Success.");

            Console.Write("Starting send and receive...  ");
            int output;
            Byte[] x = new Byte[] { 3, 12, 3 };
            Byte[] y = new Byte[3];
            jhi.SendAndRecv2(eventsSession, 10, x, ref y, out output);
            Console.WriteLine("Success.");

            Console.WriteLine("Waiting for event...  ");
            for (int i = 0; i < 60; i++)
            {
                Thread.Sleep(100);
                if (invoked)
                {
                    Console.WriteLine("Event raised! - success.");
                    break;
                }
            }

            Console.Write("UnRegistering events... ");
            jhi.UnRegisterEvents(eventsSession);
            Console.WriteLine("Success.");

            Console.Write("Closing session... ");
            jhi.CloseSession(eventsSession);
            Console.WriteLine("Success.");

            Console.Write("Uninstalling applet... ");
            jhi.Uninstall(eventServiceAppID);
            Console.WriteLine("Success.");

            return bool_To_JhiTestResult(invoked);
        }

        static void pause()
        {
#if DEBUG
            Console.WriteLine("Press any key to continue . . .");
            Console.ReadKey();
#endif
        }

        #endregion
    }
}
