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
using System.Collections;
using System.IO;
using System.Reflection;

namespace SharpTest
{
    public delegate JhiTests.JhiTestResult JHI_Function();

    class Program
    {
        static ArrayList passedTests;
        static ArrayList skippedTests;
        static ArrayList failedTests;
        static bool consoleMode = true;

        static int getCommand(string[] args)
        {
            int command = -1;
            if (args == null || args.Count() == 0)
            {
                consoleMode = false;
                printUsage();
                Console.WriteLine(">>> Please insert command number...");
                command = Console.Read();
            }
            else
            {
                try
                {
                    command = Convert.ToInt32(args[0]);
                    return command;
                }
                catch (Exception)
                {
                    printUsage();
                    Console.WriteLine(">>> Invalid argument! Please insert command <<<");
                    command = Console.Read();
                }
            }
            return command -0x30;
        }



        static int Main(string[] args)
        {
            try
            {
                int command = getCommand(args);
                int result = runCommand(command);
                pause();
                return result;
            }
            catch (Exception x)
            {
                Console.WriteLine(x.Message);
                return 1;
            }
        }

        private static void printUsage()
        {
            Console.WriteLine("\n================================  JHI SHARP TEST  ============================");
            Console.WriteLine("Usage: " + Path.GetFileName(System.Diagnostics.Process.GetCurrentProcess().MainModule.FileName) + " <Command Number>\n");
            Console.WriteLine("Available Commands:\n");
            Console.WriteLine("******************************************************************************");
            Console.WriteLine("0) Run all tests. (except sleep tests)                                        ");
            Console.WriteLine("1) Get FW Version test.                                                       ");
            Console.WriteLine("2) Send and Recieve test.                  (echo.dalp will be used)           ");
            Console.WriteLine("3) Events test.                            (EventService.dalp will be used)   ");
            Console.WriteLine("4) Delayed start test.                     (Administrator privileges required)");
            Console.WriteLine("5) Event during sleep test.                (EventService.dalp will be used)   ");
            Console.WriteLine("6) Send and Receive during sleep test.     (EventService.dalp will be used)   ");
            Console.WriteLine("7) Stop service during send and Receive.   (EventService.dalp will be used)   ");
            Console.WriteLine("******************************************************************************");
        }

        public static int runCommand(int cmd)
        {
            try
            {
                JhiTests.init();
            }
            catch (JhiException je)
            {
                Console.WriteLine("\n*** JHI init failed! " + je.Message + " - " + je.JhiRet.ToString() + "\n___///\n");
                return 1;

            }
            catch (Exception e)
            {
                Console.WriteLine("JHI init failed!");
                Console.WriteLine(e.Message);
                return 1;
            }

            passedTests = new ArrayList();
            skippedTests = new ArrayList();
            failedTests = new ArrayList();

            bool result = false;

            switch (cmd)
            {
                case 0: // run all tests 
                    result = runAllTests();
                    break;

                case 1: result = runJhiTest(JhiTests.GetFWVersion);
                    break;

                case 2: result = runJhiTest(JhiTests.SendAndReceive);
                    break;

                case 3: result = runJhiTest(JhiTests.eventsTest);
                    break;

                case 4: result = runJhiTest(JhiTests.DelayedStart);
                    break;

                case 5: result = runJhiTest(JhiTests.eventDuringSleepTest);
                    break;

                case 6: result = runJhiTest(JhiTests.SendAndReceiveDuringSleepTest);
                    break;

                case 7: result = runJhiTest(JhiTests.StopWhileSendAndReceive);
                    break;

                default:
                    Console.WriteLine(">>> Invalid Command! <<<");
                    return 1;
            }
            
            return result? 0 : 1;
        }

        private static bool runAllTests()
        {
            
            runJhiTest(JhiTests.GetFWVersion);
            runJhiTest(JhiTests.SendAndReceive);
            runJhiTest(JhiTests.eventsTest);
            runJhiTest(JhiTests.DelayedStart);
            //runJhiTest(JhiTests.StopWhileSendAndReceive);
            summarizeTest();
            if (failedTests.Count > 0)
            {
                return false;
            }
            return true;
        }

        private static void loadJhiSharp()
        {

           // Assembly.LoadFile("jhisharp");

        }

        private static void summarizeTest()
        {
            Console.WriteLine();
            if (failedTests.Count > 0)
            {
                Console.WriteLine("#####  #####   ###   #####      #####   ###  #####  #     #####  ####    #");
                Console.WriteLine("  #    #      ##       #        #      #  #    #    #     #      #   #   #");
                Console.WriteLine("  #    ###     ###     #        ###    ####    #    #     ###    #   #   #");
                Console.WriteLine("  #    #         ##    #        #      #  #    #    #     #      #   #    ");
                Console.WriteLine("  #    #####   ###     #        #      #  #  #####  ##### #####  ####    #\n");
            }
            else
            {
                Console.WriteLine("#####  #####   ###   #####     ####    ###    ###    ###   #####  ####    #");
                Console.WriteLine("  #    #      ##       #       #   #  #  #   ##     ##     #      #   #   #");
                Console.WriteLine("  #    ###     ###     #       ####   ####    ###    ###   ###    #   #   #");
                Console.WriteLine("  #    #         ##    #       #      #  #      ##     ##  #      #   #    ");
                Console.WriteLine("  #    #####   ###     #       #      #  #    ###    ###   #####  ####    #\n");
            }


            Console.WriteLine("*** Test Summary ***\n");
            Console.WriteLine("Passed tests: " + passedTests.Count);
            
            foreach (string test in passedTests)
            {
                Console.WriteLine("\t" + test);
            }

            Console.WriteLine("\nSkipped tests: " + skippedTests.Count);
            foreach (string test in skippedTests)
            {
                Console.WriteLine("\t" + test);
            }

            Console.WriteLine("\nFailed tests: " + failedTests.Count);
            foreach (string test in failedTests)
            {
                Console.WriteLine("\t" + test);
            }
            Console.WriteLine();

            
        }

        static void pause()
        {
            if (!consoleMode)
            {
                Console.WriteLine("Press any key to continue . . .");
                Console.ReadKey();
            }
        }
        static bool runJhiTest(JHI_Function fun)
        {
            try
            {
                Console.WriteLine("\n///--- Starting test " + fun.Method.Name + "...");
                JhiTests.JhiTestResult result = fun();
                if (result == JhiTests.JhiTestResult.Succeeded)
                {
                    Console.WriteLine("\n*** Test " + fun.Method.Name + " PASSED!\n___///\n");
                    passedTests.Add(fun.Method.Name);
                    return true;
                }
                else
                {
                    if (result == JhiTests.JhiTestResult.Skipped)
                    {
                        Console.WriteLine("\n*** Test " + fun.Method.Name + " FAILED!\n___///\n");
                        failedTests.Add(fun.Method.Name);
                        return false;
                    }
                    else
                    {
                        Console.WriteLine("\n*** Test " + fun.Method.Name + " SKIPPED!\n___///\n");
                        failedTests.Add(fun.Method.Name);
                        return true;
                    }
                }
            }
            catch (JhiException je)
            {
                Console.WriteLine("\n*** Test " + fun.Method.Name + " FAILED: " + je.Message + " - " + je.JhiRet.ToString() + "\n___///\n");
                failedTests.Add(fun.Method.Name);
                return false;

            }
            catch (Exception e)
            {
                Console.WriteLine("\n*** Test " + fun.Method.Name + " FAILED: " + e.Message + "\n___///\n");
                failedTests.Add(fun.Method.Name);
                return false;
            }
        }
    }
}
