PACKAGE_NAME=librdrand
PACKAGE_VERSION=1.1
CC=gcc
CFLAGS=-fPIC -O2
AR=ar

TEST_OBJ=main.o

%.o: %.c
	$(CC) $(CFLAGS) -c $<

all: bin lib 

bin: test

lib: librdrand.a

librdrand.a: rdrand.o
	$(AR) rcs $@ $?

test: $(TEST_OBJ) librdrand.a
	$(CC) $(CFLAGS) -o $@ $(TEST_OBJ) -L. -lrdrand

clean:
	rm -f *.o librdrand.a test
	
distclean: clean
	rm -f Makefile config.guess config.status config.log config.h

