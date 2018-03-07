all: all_src all_base64

all_src:
	$(MAKE) -C src
all_base64: all_src
	$(MAKE) -C base64

clean: clean_src clean_base64 clean_include
	rm -f *~ *.bak

clean_include:
	rm -f include/b64/*~

clean_src:
	$(MAKE) -C src clean;
clean_base64:
	$(MAKE) -C base64 clean;
	
distclean: clean distclean_src distclean_base64

distclean_src:
	$(MAKE) -C src distclean;
distclean_base64:
	$(MAKE) -C base64 distclean;

