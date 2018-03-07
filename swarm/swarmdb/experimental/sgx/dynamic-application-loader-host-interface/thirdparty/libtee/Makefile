#
# libtee Make File
#
INSTALL     := install
PACKAGE     := libtee
LIBOUT      := libtee.so
LIBDIR      ?= /usr/local/lib
INCDIR      ?= /usr/include/tee


LIBMEIDIR   :=  ./linux/libmei
INCLUDES    = -I$(LIBMEIDIR) -I./linux/ -I./include

CXXFLAGS    += -Wall $(INCLUDES)
CFLAGS      += -Wall -I. -fPIC -O2 $(INCLUDES)
LDFLAGS     += -Wl,-rpath=.

ifeq ($(ARCH),i386)
CFLAGS += -m32
LDFLAGS += -m32
endif


##
## Sources
##
SRCS    := linux/libteelinux.c
HEADERS := libtee.h helpers.h
OBJS    := $(SRCS:.c=.o)

all: LIBMEI_TARGET=all
all: $(LIBOUT) libmei
$(LIBOUT): $(OBJS) $(LIBMEIDIR)/mei.o
	$(CC) $(LDFLAGS) --shared $(OBJS) $(LIBMEIDIR)/mei.o -o  $@

$(LIBMEIDIR)/mei.o:
	$(MAKE) -C $(LIBMEIDIR) mei.o

clean:
	$(RM) *.o *.so *~ linux/*.o $(LIBMEIDIR)/*.o linux/*~ tags

pack: ver=$(shell git describe)
pack:
	git archive --format=tar --prefix=$(PACKAGE)-$(ver)/ HEAD | gzip > $(PACKAGE)-$(ver).tar.gz

tags: $(wildcard *.[ch])
	ctags $^

install_lib: $(LIBOUT)
	$(INSTALL) -D $(LIBOUT) $(LIBDIR)/$(LIBOUT)

install_headers: $(LIB_HEADERS)
	$(INSTALL) -m 0644 -D $^ $(INCDIR)/$^

.PHONY: clean pack install_lib install_headers libmei

