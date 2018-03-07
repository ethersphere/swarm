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

#ifndef __DBG_H__
#define __DBG_H__

#ifdef __cplusplus
extern "C" {
#endif

#include "typedefs.h"

#ifdef _WIN32
#include <tchar.h>
#else
#include <unistd.h>
#include <sys/syscall.h>
#endif

#ifdef __ANDROID__
inline static int GetCurrentThreadId(){return gettid();}
#elif defined(__linux__)
inline int GetCurrentThreadId(){return syscall(SYS_gettid);}
#endif

// Can't use 'enum class' because C code needs to use this
typedef enum JHI_LOG_LEVEL
{
	JHI_LOG_LEVEL_OFF,
	JHI_LOG_LEVEL_RELEASE,
	JHI_LOG_LEVEL_DEBUG
} JHI_LOG_LEVEL;

// Current log level in the process
extern JHI_LOG_LEVEL g_jhiLogLevel;

const char *JHIErrorToString(UINT32 retVal);
const char *TEEErrorToString(UINT32 retVal);

#define TRACE0                       	JHI_Trace
#define TRACE1(fmt,p1)               	JHI_Trace(fmt,p1)
#define TRACE2(fmt,p1,p2)            	JHI_Trace(fmt,p1,p2)
#define TRACE3(fmt,p1,p2,p3)         	JHI_Trace(fmt,p1,p2,p3)
#define TRACE4(fmt,p1,p2,p3,p4)      	JHI_Trace(fmt,p1,p2,p3,p4)
#define TRACE5(fmt,p1,p2,p3,p4,p5)   	JHI_Trace(fmt,p1,p2,p3,p4,p5)
#define TRACE6(fmt,p1,p2,p3,p4,p5,p6)	JHI_Trace(fmt,p1,p2,p3,p4,p5,p6)

#define T_TRACE1(fmt, p1)				JHI_T_Trace(fmt,p1)

#define LOG0                       		JHI_Log
#define LOG1(fmt,p1)               		JHI_Log(fmt,p1)
#define LOG2(fmt,p1,p2)            		JHI_Log(fmt,p1,p2)
#define LOG3(fmt,p1,p2,p3)         		JHI_Log(fmt,p1,p2,p3)
#define LOG4(fmt,p1,p2,p3,p4)      		JHI_Log(fmt,p1,p2,p3,p4)
#define LOG5(fmt,p1,p2,p3,p4,p5)   		JHI_Log(fmt,p1,p2,p3,p4,p5)
#define LOG6(fmt,p1,p2,p3,p4,p5,p6)		JHI_Log(fmt,p1,p2,p3,p4,p5,p6)

UINT32 JHI_Trace(const char*  Format, ... );
UINT32 JHI_T_Trace(const TCHAR* fmt, ... );
UINT32 JHI_Log(const char* Format, ...);

/*
#define LOG						JHI_Log(fmt, );


#define __DBG_PREAMBLE__
#define __DBG_POSTAMBLE__

#define __DBG_PREAMBLE__   TRACE1("===> %s: ", __FUNCTION__ ) ;
#define __DBG_POSTAMBLE__  TRACE1("<=== %s: ", __FUNCTION__ ) ;

#ifdef DEBUG

#define DISPLAY_PARA(x, l)

#ifndef ASSERT
#define ASSERT(x) \
            if (!(x)) \
            { \
                JHI_Trace ( "\n *** ASSERTION FAILED: %s line %d: " #x " ***\n",  \
                            __FILE__, __LINE__ );                         \
            }
#endif // ASSERT

#ifndef DEBUGASSERT

#define DEBUGASSERT(x) \
            if (!(x)) \
            { \
                JHI_Trace ( "\n *** ASSERTION FAILED: %s line %d: " #x " ***\n",  \
                            __FILE__, __LINE__ );                         \
            }
#endif // DEBUGASSERT

#else



#ifndef ASSERT
#define ASSERT(x)
#endif // ASSERT

#ifndef DEBUGASSERT
#define DEBUGASSERT(x)
#endif // DEBUGASSERT

#endif //  DEBUG

#ifdef DEBUG

#ifndef ASSERTMSG

#define ASSERTMSG(x,msg) \
      if (!(x))  \
      { \
         JHI_Trace ( "\n ** ASSERTION FAILED: %s line %d: " #x "\n", \
                     __FILE__, __LINE__ ) ; \
         JHI_Trace ( " *** %s\n", msg ) ; \
      }

#endif // ASSERTMSG

#else

#ifndef ASSERTMSG
#define ASSERTMSG(x,msg)
#endif // ASSERTMSG

#endif   // DEBUG
*/
#ifdef __cplusplus
};
#endif

#endif
