#ifndef _INTEL_SGX_H  
#define _INTEL_SGX_H  
  
#ifdef __cplusplus  
#include <stdio.h>
#include <iostream>
#include "Enclave_u.h"
#include "sgx_urts.h"
#include "sgx_utils/sgx_utils.h"
extern "C" {  
// int Testmain(void);
#endif  
int testMain(void);
// int Testmain(void);
/* struct Stack
{
	int * base;
	int * top;
	int size;
};
extern void myTest();
extern void myTest(struct Stack stk);  
extern void myTest2(struct Stack* stk);  
// extern void myTest3(struct Stack& stk);  
extern void TEST();   */
  
#ifdef __cplusplus  
}  
#endif  
  
#endif  

