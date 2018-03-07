#
INSTALL := install
PACKAGE := libmei
HEADERS := libmei.h
SOURCES := mei.c
#
MEI := mei
LIBS := lib$(MEI).so

all: $(LIBS)

INCLUDES := -I. -I./include
CXXFLAGS=-Wall
CFLAGS += -Wall -ggdb $(INCLUDES) -fPIC -O2
LDFLAGS += -Wl,-rpath=.

ifeq ($(ARCH),i386)
CFLAGS += -m32
LDFLAGS += -m32
endif

LIBDIR ?= /usr/local/lib
INCDIR ?= /usr/include/$(MEI)

lib%.so: %.o
	$(CC) $(LDFLAGS) --shared $^ -o $@

clean:
	$(RM) $(LIBS) *.o

dist-clean: TARGET=clean
dist-clean: doc clean
	$(RM) tags

pack: ver=$(shell git describe)
pack:
	git archive --format=tar --prefix=$(PACKAGE)-$(ver)/ HEAD | gzip > $(PACKAGE)-$(ver).tar.gz

pack-doc: ver=$(shell git describe)
pack-doc: doc
	tar -czf $(PACKAGE)-doc-$(ver).tar.gz doc/html

tags: $(wildcard *.[ch])
	ctags $^

install_lib: $(LIBS)
	$(INSTALL) -D $^ $(LIBDIR)/$^

install_headers: $(HEADERS)
	$(INSTALL) -m 0644 -D $^ $(INCDIR)/$^

doc: $(HEADERS) $(SOURCES)
	$(MAKE) -C doc $(TARGET) SRCDIR=$(PWD)

.PHONY: clean tags doc dist-clean
